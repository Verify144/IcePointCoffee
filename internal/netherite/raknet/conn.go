package raknet

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Verify144/IcePointCoffee/internal/netherite/crypto"
)

// Conn 表示一个 Raknet 连接。
type Conn struct {
	conn     *net.UDPConn
	addr     *net.UDPAddr
	mtu      uint16
	state    State
	guid     int64
	clientGUID int64

	// 加密
	ecdh          *crypto.ECDH
	sendCipher    *crypto.StreamCipher
	recvCipher    *crypto.StreamCipher
	encrypted     atomic.Bool

	// 连接参数
	maxTransferUnit uint16

	// 序列号
	sendReliableIdx uint32
	recvReliableIdx uint32
	sendSeqIdx     uint32
	recvSeqIdx     uint32

	// 重传
	pendingPackets  map[uint32]*Frame // reliable idx → frame
	ackQueue       []uint32
	nakQueue       []uint32

	// 分片重组
	splitFragments map[uint32]map[uint32]*Frame // splitID → (splitIdx → frame)

	// 接收缓冲（帧队列）
	recvQueue  chan *Frame
	sendQueue  chan *Frame

	// 关闭
	closeOnce sync.Once
	closed    atomic.Bool

	mu sync.Mutex
}

// Dial 建立 Raknet 连接。
func Dial(ctx context.Context, addr string, mtu uint16) (*Conn, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, fmt.Errorf("resolve udp addr: %w", err)
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		return nil, fmt.Errorf("dial udp: %w", err)
	}

	// 设置读写超时
	conn.SetDeadline(time.Now().Add(30 * time.Second))
	defer conn.SetDeadline(time.Time{})

	if mtu == 0 {
		mtu = MTUDefault
	}
	if mtu > 1400 {
		mtu = 1400
	}

	// 生成客户端 GUID
	var guid int64
	binary.Read(rand.Reader, binary.LittleEndian, &guid)

	c := &Conn{
		conn:       conn,
		addr:       udpAddr,
		mtu:        mtu,
		guid:       guid,
		clientGUID: guid,
		state:      StateUnconnected,
		pendingPackets: make(map[uint32]*Frame),
		splitFragments: make(map[uint32]map[uint32]*Frame),
		recvQueue:  make(chan *Frame, 256),
		sendQueue:  make(chan *Frame, 256),
	}

	// 生成 ECDH 密钥对
	ecdh, err := crypto.NewECDH()
	if err != nil {
		c.Close()
		return nil, fmt.Errorf("generate ecdh: %w", err)
	}
	c.ecdh = ecdh
	c.ackQueue = make([]uint32, 0)
	c.nakQueue = make([]uint32, 0)

	// Step 1: OpenConnectionRequest1
	if err := c.step1(ctx); err != nil {
		c.Close()
		return nil, fmt.Errorf("step1: %w", err)
	}

	// Step 2: OpenConnectionRequest2
	if err := c.step2(ctx); err != nil {
		c.Close()
		return nil, fmt.Errorf("step2: %w", err)
	}

	// Step 3: ConnectionRequest
	if err := c.step3(ctx); err != nil {
		c.Close()
		return nil, fmt.Errorf("step3: %w", err)
	}

	// Step 4: ECDH 加密握手
	if err := c.step4(ctx); err != nil {
		// 加密握手失败不一定要断开（部分服务器不支持加密）
		log.Printf("[Raknet] encrypted handshake skipped: %v", err)
	}

	c.state = StateConnected

	// 启动读写 goroutine
	go c.readLoop()
	go c.writeLoop()
	go c.heartbeatLoop()

	return c, nil
}

// step1 发送 OpenConnectionRequest1 并处理 Reply1
func (c *Conn) step1(ctx context.Context) error {
	req := &OpenConnectionRequest1{
		MTU:            c.mtu,
		ProtocolVersion: ProtocolVersion,
	}
	_, err := c.conn.Write(req.Encode())
	if err != nil {
		return err
	}

	// 等待 Reply1
	buf := make([]byte, 1500)
	c.conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	n, _, err := c.conn.ReadFromUDP(buf)
	if err != nil {
		return fmt.Errorf("wait reply1: %w", err)
	}
	c.conn.SetReadDeadline(time.Time{})

	if n < 2 || buf[0] != IDOpenConnectionReply1 {
		return errors.New("invalid reply1")
	}

	// Reply1: 跳过 PacketID(1) + Magic(16) = 17
	offset := 17
	serverGUID := int64(binary.LittleEndian.Uint64(buf[offset : offset+8]))
	_ = serverGUID // 记录服务端 GUID
	offset += 8
	mtuReply := binary.LittleEndian.Uint16(buf[offset : offset+2])
	_ = mtuReply

	c.maxTransferUnit = c.mtu
	return nil
}

// step2 发送 OpenConnectionRequest2 并处理 Reply2
func (c *Conn) step2(ctx context.Context) error {
	req := &OpenConnectionRequest2{
		MTU:      c.mtu,
		ClientGUID: c.clientGUID,
	}
	_, err := c.conn.Write(req.Encode())
	if err != nil {
		return err
	}

	buf := make([]byte, 1500)
	c.conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	n, _, err := c.conn.ReadFromUDP(buf)
	if err != nil {
		return fmt.Errorf("wait reply2: %w", err)
	}
	c.conn.SetReadDeadline(time.Time{})

	if n < 2 || buf[0] != IDOpenConnectionReply2 {
		return errors.New("invalid reply2")
	}

	return nil
}

// step3 发送 ConnectionRequest 并等待 NewConnection
func (c *Conn) step3(ctx context.Context) error {
	// 构造 ConnectionRequest 包
	pkt := c.packSystemPacket(IDConnectionRequest, func(b []byte) {
		offset := 0
		b[offset] = IDConnectionRequest
		offset++
		// timestamp
		ts := time.Now().UnixMilli()
		binary.LittleEndian.PutUint64(b[offset:], uint64(ts))
		offset += 8
		// client GUID
		binary.LittleEndian.PutUint64(b[offset:], uint64(c.clientGUID))
		offset += 8
		// use security = false
		b[offset] = 0
		offset++
		// enabled FTS = true (1 byte, set to 0)
		b[offset] = 0
	})

	_, err := c.conn.Write(pkt)
	if err != nil {
		return err
	}

	// 等待 NewConnection
	buf := make([]byte, 1500)
	c.conn.SetReadDeadline(time.Now(). Add(10 * time.Second))
	n, fromAddr, err := c.conn.ReadFromUDP(buf)
	if err != nil {
		return fmt.Errorf("wait new connection: %w", err)
	}
	c.conn.SetReadDeadline(time.Time{})

	// 检查来源（可能回复到不同端口）
	c.addr = fromAddr

	// 解析包
	packetID := buf[0]
	if packetID == IDConnectionResponse {
		// 连接成功
		return nil
	} else if packetID == IDIncompatibleProtocol {
		return errors.New("incompatible protocol")
	} else if packetID == IDDisconnect {
		return errors.New("server disconnected")
	}

	// 可能是其他帧，先缓存
	if n > 0 {
		_ = c.handleRawPacket(buf[:n])
	}

	return nil
}

// step4 ECDH 加密握手
func (c *Conn) step4(ctx context.Context) error {
	// 1. 生成 ECDH 密钥对
	if c.ecdh == nil {
		return errors.New("ecdh not initialized")
	}

	// 2. 发送 Encrypted Handshake（带公钥）
	handshake := &crypto.EncryptedHandshake{
		Cookie:    []byte{0x00, 0x00, 0x00, 0x00},
		PublicKey: c.ecdh.PublicKeyBytes(),
		Challenge: crypto.RandomBytes(16),
	}

	// 3. 接收服务器响应
	_ = handshake
	// 网易服务器可能直接接受连接不返回加密握手
	// 简化：暂不启用加密发送/接收（保持明文）
	// 实际生产中，接收后会派生共享密钥、计算流加密器
	return errors.New("encrypted mode not yet active")
}

// packSystemPacket 将系统包打包为帧
func (c *Conn) packSystemPacket(id byte, fill func([]byte)) []byte {
	payload := make([]byte, 64)
	fill(payload)
	payload[0] = id

	frame := &Frame{
		Flags: FramePriorityImmediate,
		Data:  payload,
	}
	// 直接发送（不走重传队列）
	return encodeFrameToPacket(frame)
}

// encodeFrameToPacket 将帧编码为发送包
func encodeFrameToPacket(frame *Frame) []byte {
	data, _ := frame.Encode()
	return data
}

// readLoop 持续读取 UDP 数据
func (c *Conn) readLoop() {
	buf := make([]byte, 2048)
	for !c.closed.Load() {
		c.conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		n, _, err := c.conn.ReadFromUDP(buf)
		if c.closed.Load() {
			return
		}
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			log.Printf("read error: %v", err)
			break
		}
		if n > 0 {
			c.handleRawPacket(buf[:n])
		}
	}
}

// handleRawPacket 处理原始 UDP 包
func (c *Conn) handleRawPacket(data []byte) error {
	if len(data) == 0 {
		return nil
	}

	packetID := data[0]

	// 系统包（不需要帧层）
	switch packetID {
	case IDConnectionRequest:
		return c.handleConnectionRequest(data)
	case IDConnectionResponse:
		return c.handleConnectionResponse(data)
	case IDNewConnection:
		return c.handleNewConnection(data)
	case IDDisconnect:
		return c.handleDisconnect(data)
	case IDIncompatibleProtocol:
		return errors.New("incompatible protocol")
	case IDUnconnectedPong:
		return nil // ping pong
	}

	// 游戏帧：先解密（如果已启用加密）
	decrypted := data
	if c.encrypted.Load() && c.recvCipher != nil {
		decrypted = c.recvCipher.DecryptCopy(data)
	}

	offset := 0
	for offset < len(decrypted) {
		frame, n, err := DecodeFrame(decrypted[offset:])
		if err != nil {
			break
		}
		offset += n

		if err := c.handleFrame(frame); err != nil {
			log.Printf("handle frame error: %v", err)
		}
	}

	return nil
}

// handleFrame 处理单个帧
func (c *Conn) handleFrame(frame *Frame) error {
	// 分片
	if frame.IsSplit {
		return c.handleSplitFrame(frame)
	}

	// 可靠帧：发送 ACK
	if frame.Flags&FlagReliable != 0 {
		c.ackQueue = append(c.ackQueue, frame.ReliableIdx)
	}

	// 放入接收队列
	select {
	case c.recvQueue <- frame:
	default:
		// 队列满，丢帧
	}
	return nil
}

// handleSplitFrame 处理分片帧
func (c *Conn) handleSplitFrame(frame *Frame) error {
	splitID := frame.SplitID
	if _, ok := c.splitFragments[splitID]; !ok {
		c.splitFragments[splitID] = make(map[uint32]*Frame)
	}
	c.splitFragments[splitID][frame.SplitIndex] = frame

	// 检查是否完整
	fragments := c.splitFragments[splitID]
	if uint32(len(fragments)) == frame.SplitCount {
		// 组装完整数据
		complete := make([]byte, 0)
		for i := uint32(0); i < frame.SplitCount; i++ {
			f := fragments[i]
			complete = append(complete, f.Data...)
		}
		// 清除分片
		delete(c.splitFragments, splitID)

		// 模拟完整帧处理
		reassembled := &Frame{
			Flags: frame.Flags &^ byte(0x0F), // 清除分片标志
			Data:  complete,
		}
		if reassembled.Flags&FlagReliable != 0 {
			c.ackQueue = append(c.ackQueue, frame.ReliableIdx)
		}
		select {
		case c.recvQueue <- reassembled:
		default:
		}
	}
	return nil
}

// handleConnectionRequest 处理连接请求
func (c *Conn) handleConnectionRequest(data []byte) error {
	// 不在客户端侧处理
	return nil
}

// handleConnectionResponse 处理连接响应
func (c *Conn) handleConnectionResponse(data []byte) error {
	// 不在客户端侧处理
	return nil
}

// handleNewConnection 处理新连接
func (c *Conn) handleNewConnection(data []byte) error {
	return nil
}

// handleDisconnect 处理断开
func (c *Conn) handleDisconnect(data []byte) error {
	c.Close()
	return errors.New("server disconnected")
}

// writeLoop 处理发送队列和 ACK
func (c *Conn) writeLoop() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for !c.closed.Load() {
		select {
		case <-ticker.C:
			c.flushAcks()
		case frame := <-c.sendQueue:
			c.sendFrame(frame)
		}
	}
}

// heartbeatLoop 心跳循环
func (c *Conn) heartbeatLoop() {
	ticker := time.NewTicker(HeartbeatInterval * time.Millisecond)
	defer ticker.Stop()

	for !c.closed.Load() {
		<-ticker.C
		if c.closed.Load() {
			return
		}
		// 发送心跳包
		c.mu.Lock()
		idx := c.sendReliableIdx
		c.sendReliableIdx++
		c.mu.Unlock()

		frame := &Frame{
			Flags:       FlagReliable,
			ReliableIdx: idx,
			Data:        []byte{0x00}, // heartbeat
		}
		c.sendFrame(frame)
	}
}

// flushAcks 批量发送 ACK
func (c *Conn) flushAcks() {
	c.mu.Lock()
	if len(c.ackQueue) == 0 {
		c.mu.Unlock()
		return
	}
	acks := c.ackQueue
	c.ackQueue = make([]uint32, 0)
	c.mu.Unlock()

	// ACK 包
	pkt := c.encodeACK(acks)
	_, err := c.conn.WriteToUDP(pkt, c.addr)
	if err != nil {
		log.Printf("send ack error: %v", err)
	}
}

// encodeACK 编码 ACK 包
func (c *Conn) encodeACK(acks []uint32) []byte {
	if len(acks) == 0 {
		return nil
	}
	buf := make([]byte, 0, 8+len(acks)*3)
	buf = append(buf, 0xC0) // ACK
	buf = append(buf, byte(len(acks)&0xFF), byte((len(acks)>>8)&0xFF)) // count
	for _, idx := range acks {
		buf = append(buf, byte(idx), byte(idx>>8), byte(idx>>16))
	}
	return buf
}

// sendFrame 发送帧
func (c *Conn) sendFrame(frame *Frame) {
	c.mu.Lock()
	if frame.Flags&FlagReliable != 0 {
		frame.ReliableIdx = c.sendReliableIdx
		c.sendReliableIdx++
	}
	c.mu.Unlock()

	data, err := frame.Encode()
	if err != nil {
		log.Printf("encode frame error: %v", err)
		return
	}

	packet := make([]byte, len(data)+1)
	packet[0] = 0x80 // Game packet
	copy(packet[1:], data)

	// 加密（如果已启用）
	if c.encrypted.Load() && c.sendCipher != nil {
		encrypted := c.sendCipher.EncryptCopy(packet)
		_, err = c.conn.WriteToUDP(encrypted, c.addr)
	} else {
		_, err = c.conn.WriteToUDP(packet, c.addr)
	}
	if err != nil {
		log.Printf("send frame error: %v", err)
	}
}

// Send 发送游戏数据帧
func (c *Conn) Send(data []byte) error {
	if c.closed.Load() {
		return errors.New("connection closed")
	}
	frame := &Frame{
		Flags: FramePriorityNormal,
		Data:  data,
	}
	select {
	case c.sendQueue <- frame:
		return nil
	default:
		return errors.New("send queue full")
	}
}

// Recv 接收游戏数据帧
func (c *Conn) Recv() (*Frame, error) {
	select {
	case frame := <-c.recvQueue:
		return frame, nil
	case <-time.After(5 * time.Second):
		return nil, errors.New("recv timeout")
	}
}

// RecvNonBlock 非阻塞接收
func (c *Conn) RecvNonBlock() (*Frame, error) {
	select {
	case frame := <-c.recvQueue:
		return frame, nil
	default:
		return nil, nil
	}
}

// Close 关闭连接
func (c *Conn) Close() error {
	c.closeOnce.Do(func() {
		c.closed.Store(true)
		if c.conn != nil {
			// 发送 Disconnect 包
			disconnect := []byte{IDDisconnect}
			c.conn.WriteToUDP(disconnect, c.addr)
			c.conn.Close()
		}
		close(c.recvQueue)
		close(c.sendQueue)
	})
	return nil
}

// EnableEncryption 启用加密。
// 接收 ECDH 共享密钥后调用此方法初始化流加密器。
func (c *Conn) EnableEncryption(sendCipher, recvCipher *crypto.StreamCipher) {
	c.sendCipher = sendCipher
	c.recvCipher = recvCipher
	c.encrypted.Store(true)
}

// IsEncrypted 返回是否已加密
func (c *Conn) IsEncrypted() bool {
	return c.encrypted.Load()
}

// RemoteAddr 返回远端地址
func (c *Conn) RemoteAddr() *net.UDPAddr {
	return c.addr
}

// WriteToUDP 发送原始 UDP 数据
func (c *Conn) WriteToUDP(data []byte, addr *net.UDPAddr) (int, error) {
	if c.closed.Load() {
		return 0, errors.New("connection closed")
	}
	return c.conn.WriteToUDP(data, c.addr)
}

// SetReadDeadline 设置读超时
func (c *Conn) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}

// State 返回连接状态
func (c *Conn) State() State {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.state
}

// GUID 返回服务端 GUID
func (c *Conn) GUID() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.guid
}

// LocalGUID 返回本地 GUID
func (c *Conn) LocalGUID() int64 {
	return c.clientGUID
}
