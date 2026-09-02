// Package mc 提供 Minecraft Bedrock 连接管理。
// 整合 Raknet 传输层、FB 认证、登录握手和游戏协议。
package mc

import (
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/Verify144/IcePointCoffee/internal/netherite/auth"
	"github.com/Verify144/IcePointCoffee/internal/netherite/crypto"
	"github.com/Verify144/IcePointCoffee/internal/netherite/protocol"
	"github.com/Verify144/IcePointCoffee/internal/netherite/raknet"
)

// Client MC 客户端。
type Client struct {
	authClient  *auth.Client
	fbToken     string
	serverCode  string
	serverPass  string
	playerName   string
	address      string
	port         uint16

	conn          *raknet.Conn
	loginTokens   protocol.LoginTokens
	authResp      *auth.LoginResponse
	ecdhKey       *crypto.ECDH

	state        ConnState
	connectedAt  time.Time

	// 事件
	mu         sync.RWMutex
	onText     func(*protocol.Text)
	onCommand  func(*protocol.CommandOutput)
	onStatus   func(int32)

	wg sync.WaitGroup
}

// ConnState 连接状态
type ConnState uint8

const (
	StateInit         ConnState = iota
	StateAuth         // FB 认证完成
	StateRaknet       // Raknet 连接建立
	StateLoginSent    // Login 包已发送
	StatePlaying      // 进入游戏
	StateClosed
)

// Options 连接选项
type Options struct {
	FBToken     string // FB Master Token
	ServerCode  string // 房间号
	ServerPass  string // 房间密码（可空）
	PlayerName  string // 玩家名称
	AuthServer  string // 认证服务器
}

// NewClient 创建 MC 客户端
func NewClient(opts *Options) (*Client, error) {
	if opts.FBToken == "" {
		return nil, fmt.Errorf("fb_token is required")
	}
	if opts.ServerCode == "" {
		return nil, fmt.Errorf("server_code is required")
	}
	if opts.PlayerName == "" {
		return nil, fmt.Errorf("player_name is required")
	}

	authOpts := &auth.Options{
		AuthServer: opts.AuthServer,
	}
	authClient, err := auth.NewClient(authOpts)
	if err != nil {
		return nil, fmt.Errorf("create auth client: %w", err)
	}

	// 生成 X25519 ECDH 密钥对
	ecdhKey, err := crypto.NewECDH()
	if err != nil {
		return nil, fmt.Errorf("generate ecdh key: %w", err)
	}

	return &Client{
		authClient: authClient,
		fbToken:   opts.FBToken,
		serverCode: opts.ServerCode,
		serverPass: opts.ServerPass,
		playerName: opts.PlayerName,
		state:      StateInit,
		ecdhKey:   ecdhKey,
	}, nil
}

// Connect 连接到租赁服。
// 流程：FB认证 → Raknet连接 → Login握手 → 进服验证 → Playing
func (c *Client) Connect(ctx context.Context) error {
	// Step 1: FB 认证
	if err := c.stepAuth(ctx); err != nil {
		return fmt.Errorf("auth failed: %w", err)
	}

	// Step 2: Raknet 连接
	if err := c.stepRaknet(ctx); err != nil {
		return fmt.Errorf("raknet connect failed: %w", err)
	}

	// Step 3: Login 握手
	if err := c.stepLogin(ctx); err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	// Step 4: 进服后握手（网易 PyRpc）
	if err := c.stepPostLogin(); err != nil {
		return fmt.Errorf("post-login handshake failed: %w", err)
	}

	// Step 5: 处理事件循环
	go c.eventLoop()

	c.mu.Lock()
	c.state = StatePlaying
	c.connectedAt = time.Now()
	c.mu.Unlock()

	return nil
}

// stepAuth FB 认证
func (c *Client) stepAuth(ctx context.Context) error {
	// X25519 公钥（auth server 主要用 chain_data，public key 可选）
	publicKey := c.ecdhKey.PublicKeyBytes()

	loginResp, err := c.authClient.Login(ctx, &auth.LoginRequest{
		LoginToken:     c.fbToken,
		ServerCode:     c.serverCode,
		ServerPasscode: c.serverPass,
		ClientPublicKey: publicKey,
	})
	if err != nil {
		return fmt.Errorf("auth login: %w", err)
	}

	// 解析 chain/identity
	tokens, err := protocol.BuildLoginChain(loginResp.IdentityData, loginResp.ChainInfo)
	if err != nil {
		return fmt.Errorf("build login chain: %w", err)
	}
	// 覆盖显示名
	if tokens.ExtraData.DisplayName == "" {
		tokens.ExtraData.DisplayName = c.playerName
		tokens.ExtraData.Identity = c.playerName
	}

	c.loginTokens = tokens
	c.authResp = loginResp
	c.address = loginResp.IPAddress

	// 从地址提取端口
	// 格式: ip:port
	for i := len(c.address) - 1; i >= 0; i-- {
		if c.address[i] == ':' {
			var port int
			fmt.Sscanf(c.address[i+1:], "%d", &port)
			c.port = uint16(port)
			c.address = c.address[:i]
			break
		}
	}
	if c.port == 0 {
		c.port = 19132
	}

	c.mu.Lock()
	c.state = StateAuth
	c.mu.Unlock()

	log.Printf("[MC] Auth OK: uid=%d ip=%s", loginResp.UID, loginResp.IPAddress)
	return nil
}

// stepRaknet 建立 Raknet 连接
func (c *Client) stepRaknet(ctx context.Context) error {
	addr := fmt.Sprintf("%s:%d", c.address, c.port)
	conn, err := raknet.Dial(ctx, addr, 0)
	if err != nil {
		return fmt.Errorf("raknet dial: %w", err)
	}

	c.conn = conn
	c.mu.Lock()
	c.state = StateRaknet
	c.mu.Unlock()

	log.Printf("[MC] Raknet connected: %s", addr)
	return nil
}

// stepLogin 发送 Login 包并处理响应
func (c *Client) stepLogin(ctx context.Context) error {
	// 构造 Login 包
	loginData, err := protocol.EncodeLogin(protocol.ProtocolVersion, c.loginTokens)
	if err != nil {
		return fmt.Errorf("encode login: %w", err)
	}

	// 包装为帧
	frame := &raknet.Frame{
		Flags:       raknet.FlagReliable,
		ReliableIdx: 0,
		Data:        loginData,
	}
	encoded, _ := frame.Encode()

	packet := make([]byte, len(encoded)+1)
	packet[0] = 0x80
	copy(packet[1:], encoded)

	_, err = c.conn.WriteToUDP(packet, c.conn.RemoteAddr())
	if err != nil {
		return fmt.Errorf("send login: %w", err)
	}

	// 等待 PlayStatus 响应
	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		c.conn.SetReadDeadline(deadline)
		frame, err := c.conn.Recv()
		if err != nil {
			return fmt.Errorf("wait login response: %w", err)
		}
		if len(frame.Data) == 0 {
			continue
		}

		packetID := frame.Data[0]
		if packetID == protocol.IDPlayStatus {
			status := binary.LittleEndian.Uint32(frame.Data[1:5])
			if status == protocol.PlayStatusLoginSuccess {
				log.Printf("[MC] Login success!")
				return nil
			}
			return fmt.Errorf("login failed with status: %d", status)
		}

		if packetID == protocol.IDDisconnect {
			return fmt.Errorf("server disconnected during login")
		}

		// 继续等待
	}

	return fmt.Errorf("login timeout")
}

// stepPostLogin 进服后握手（网易特殊包）
func (c *Client) stepPostLogin() error {
	// 发送 ClientCacheStatus
	c.SendRaw(protocol.EncodeClientCacheStatus(false))

	// 发送 LOGIN_UID
	c.SendRaw(protocol.EncodeNeteaseJson("LOGIN_UID", "", fmt.Sprintf("%d", c.authResp.UID)))

	// 发送 PyRpc 握手序列（网易特殊）
	c.sendPyRpc("SyncUsingMod", c.buildModList())
	c.sendPyRpc("ClientLoadAddonsFinishedFromGac", []any{})
	c.sendPyRpc("arenaGamePlayerFinishLoad", []any{})
	c.sendPyRpc("ModEventC2S", []any{
		"Minecraft",
		"vipEventSystem",
		"PlayerUiInit",
		fmt.Sprintf("%d", 0), // entity unique id
	})
	c.sendPyRpc("ClientInitUIFinishedEventFromGac", []any{})

	log.Printf("[MC] Post-login handshake sent")
	return nil
}

// buildModList 构建 mod 列表
func (c *Client) buildModList() []any {
	modList := []string{}
	outfitInfo := c.authResp.OutfitInfo
	if outfitInfo == nil {
		outfitInfo = make(map[string]any)
	}
	for uuid := range outfitInfo {
		modList = append(modList, uuid)
	}
	return []any{
		modList,
		"", // skinID
		"", // skinItemID
		true,
		outfitInfo,
	}
}

// sendPyRpc 发送 PyRpc 包
func (c *Client) sendPyRpc(method string, args []any) {
	data := protocol.EncodePyRpc(method, args, protocol.PyRpcOperationTypeSend)
	c.SendFrame(data)
}

// sendFrame 发送游戏帧
func (c *Client) SendFrame(data []byte) {
	frame := &raknet.Frame{
		Flags:       raknet.FlagReliable,
		ReliableIdx: 0,
		Data:        data,
	}
	encoded, _ := frame.Encode()
	packet := make([]byte, len(encoded)+1)
	packet[0] = 0x80
	copy(packet[1:], encoded)
	c.conn.WriteToUDP(packet, c.conn.RemoteAddr())
}

// sendRaw 发送原始数据
func (c *Client) SendRaw(data []byte) {
	c.SendFrame(data)
}

// eventLoop 事件循环
func (c *Client) eventLoop() {
	c.wg.Add(1)
	defer c.wg.Done()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		c.mu.RLock()
		state := c.state
		c.mu.RUnlock()

		if state == StateClosed {
			return
		}

		frame, err := c.conn.RecvNonBlock()
		if err != nil {
			// 超时，正常继续
			<-ticker.C
			continue
		}
		if frame == nil {
			<-ticker.C
			continue
		}

		c.handlePacket(frame.Data)
	}
}

// handlePacket 处理游戏包
func (c *Client) handlePacket(data []byte) {
	if len(data) == 0 {
		return
	}
	packetID := data[0]

	switch packetID {
	case protocol.IDText:
		text, err := c.parseText(data)
		if err == nil {
			c.mu.RLock()
			fn := c.onText
			c.mu.RUnlock()
			if fn != nil {
				fn(text)
			}
		}

	case protocol.IDCommandOutput:
		output, err := c.parseCommandOutput(data)
		if err == nil {
			c.mu.RLock()
			fn := c.onCommand
			c.mu.RUnlock()
			if fn != nil {
				fn(output)
			}
		}

	case protocol.IDPlayStatus:
		if len(data) >= 5 {
			status := binary.LittleEndian.Uint32(data[1:5])
			c.mu.RLock()
			fn := c.onStatus
			c.mu.RUnlock()
			if fn != nil {
				fn(int32(status))
			}
		}

	case protocol.IDPyRpc:
		c.handlePyRpc(data)

	case protocol.IDDisconnect:
		log.Printf("[MC] Server disconnected")
		c.Close()
	}
}

// handlePyRpc 处理网易 RPC
func (c *Client) handlePyRpc(data []byte) {
	rpc, err := protocol.DecodePyRpc(data)
	if err != nil {
		return
	}
	if len(rpc.Value) < 1 {
		return
	}
	methodName, ok := rpc.Value[0].(string)
	if !ok {
		return
	}
	log.Printf("[MC] PyRpc: %s", methodName)
	// 根据 method 处理
}

// parseText 解析 Text 包
func (c *Client) parseText(data []byte) (*protocol.Text, error) {
	br := protocol.NewReader(data[1:])
	text := &protocol.Text{}
	text.Type = protocol.TextType(br.ReadByte())
	text.NeedsTranslation = br.ReadBool()
	text.SourceName = br.ReadString()
	text.XUID = br.ReadString()
	text.Message = br.ReadString()

	paramCount := br.ReadVarint()
	text.Parameters = make([]string, paramCount)
	for i := 0; i < int(paramCount); i++ {
		text.Parameters[i] = br.ReadString()
	}
	return text, nil
}

// parseCommandOutput 解析命令输出
func (c *Client) parseCommandOutput(data []byte) (*protocol.CommandOutput, error) {
	br := protocol.NewReader(data[1:])
	output := &protocol.CommandOutput{}

	// Origin
	output.Origin.Origin = protocol.CommandOriginType(br.ReadVarint())
	for i := range output.Origin.UUID {
		output.Origin.UUID[i] = br.ReadByte()
	}
	output.Origin.RequestID = br.ReadString()
	output.Origin.PlayerID = br.ReadInt64()

	output.OutputType = br.ReadByte()
	output.SuccessCount = br.ReadInt32()

	count := br.ReadVarint()
	output.Messages = make([]protocol.CommandOutputMessage, count)
	for i := 0; i < int(count); i++ {
		output.Messages[i].Success = br.ReadBool()
		output.Messages[i].MessageID = br.ReadString()
		pc := br.ReadVarint()
		output.Messages[i].Parameters = make([]string, pc)
		for j := 0; j < int(pc); j++ {
			output.Messages[i].Parameters[j] = br.ReadString()
		}
	}
	return output, nil
}

// SendCommand 发送命令并等待响应
func (c *Client) SendCommand(ctx context.Context, cmd string) (*protocol.CommandOutput, error) {
	// WS 方式
	data := protocol.EncodeWSCmd(cmd)
	c.SendFrame(data)

	// 等待命令输出
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		frame, err := c.conn.Recv()
		if err != nil {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
			}
			continue
		}
		if len(frame.Data) > 0 && frame.Data[0] == protocol.IDCommandOutput {
			return c.parseCommandOutput(frame.Data)
		}
	}
	return nil, fmt.Errorf("command timeout")
}

// SendChat 发送聊天消息
func (c *Client) SendChat(msg string) error {
	data := c.encodeChat(msg)
	c.SendFrame(data)
	return nil
}

// encodeChat 编码聊天消息
func (c *Client) encodeChat(msg string) []byte {
	buf := new(protocol.Writer)
	buf.WriteByte(protocol.IDText)
	buf.WriteByte(byte(protocol.TextTypeChat)) // type
	buf.WriteBool(false)                       // needs translation
	buf.WriteString("")                         // source name
	buf.WriteString("")                         // xuid
	buf.WriteString(msg)                        // message
	buf.WriteVarint(0)                          // param count
	return buf.Bytes()
}

// SetOnText 设置文本回调
func (c *Client) SetOnText(fn func(*protocol.Text)) {
	c.mu.Lock()
	c.onText = fn
	c.mu.Unlock()
}

// SetOnCommand 设置命令输出回调
func (c *Client) SetOnCommand(fn func(*protocol.CommandOutput)) {
	c.mu.Lock()
	c.onCommand = fn
	c.mu.Unlock()
}

// SetOnStatus 设置状态回调
func (c *Client) SetOnStatus(fn func(int32)) {
	c.mu.Lock()
	c.onStatus = fn
	c.mu.Unlock()
}

// Close 关闭连接
func (c *Client) Close() error {
	c.mu.Lock()
	if c.state == StateClosed {
		c.mu.Unlock()
		return nil
	}
	c.state = StateClosed
	c.mu.Unlock()

	if c.conn != nil {
		c.conn.Close()
	}
	c.wg.Wait()
	return nil
}

// IsConnected 返回是否已连接
func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.state == StatePlaying || c.state == StateLoginSent
}

// RemoteAddr 返回 Raknet 远端地址
func (c *Client) RemoteAddr() string {
	return fmt.Sprintf("%s:%d", c.address, c.port)
}
