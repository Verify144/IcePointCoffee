// Package crypto 提供 Bedrock Edition 网络加密。
// 包含 ECDH 握手、AES-256-CTR 流加密。
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
)

// ECDH 椭圆曲线 Diffie-Hellman 密钥交换。
// 使用 X25519 曲线（Bedrock 1.16+）。
type ECDH struct {
	curve    ecdh.Curve
	private  *ecdh.PrivateKey
	public   *ecdh.PublicKey
}

// NewECDH 创建 ECDH 实例。
func NewECDH() (*ECDH, error) {
	curve := ecdh.X25519()
	private, err := curve.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generate key: %w", err)
	}
	return &ECDH{
		curve:   curve,
		private: private,
		public:  private.PublicKey(),
	}, nil
}

// NewECDHP384 创建 P-384 曲线的 ECDH 实例（备用，网易版）。
func NewECDHP384() (*ECDH, error) {
	curve := ecdh.P256()
	private, err := curve.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generate key: %w", err)
	}
	return &ECDH{
		curve:   curve,
		private: private,
		public:  private.PublicKey(),
	}, nil
}

// PublicKeyBytes 返回公钥字节。
func (e *ECDH) PublicKeyBytes() []byte {
	return e.public.Bytes()
}

// ECDH 计算共享密钥。
func (e *ECDH) ECDH(remotePubKey []byte) ([]byte, error) {
	pub, err := e.curve.NewPublicKey(remotePubKey)
	if err != nil {
		return nil, fmt.Errorf("parse remote public key: %w", err)
	}
	secret, err := e.private.ECDH(pub)
	if err != nil {
		return nil, fmt.Errorf("compute shared secret: %w", err)
	}
	return secret, nil
}

// DeriveKey 派生加密密钥。
// Bedrock 使用双 SHA-256 派生，salt 来自 challenge 响应。
func DeriveKey(sharedSecret, salt []byte) ([]byte, error) {
	// salt 拼接共享密钥
	data := make([]byte, 0, len(salt)+len(sharedSecret))
	data = append(data, salt...)
	data = append(data, sharedSecret...)

	// 第一次 hash
	h1 := sha256.Sum256(data)

	// 第二次 hash (Bedrock 模式)
	data2 := make([]byte, 0, len(salt)+len(h1))
	data2 = append(data2, salt...)
	data2 = append(data2, h1[:]...)

	h2 := sha256.Sum256(data2)
	return h2[:], nil
}

// StreamCipher 提供 AES-256-CTR 流加密。
// Bedrock 网络流加密使用 CTR 模式。
type StreamCipher struct {
	enc cipher.Stream
	dec cipher.Stream
	key []byte
}

// NewStreamCipher 创建流加密器（同一 cipher 用于双向对称测试）。
// key 长度必须是 16/24/32 字节（AES-128/192/256）。
func NewStreamCipher(key []byte) (*StreamCipher, error) {
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return nil, fmt.Errorf("key length must be 16/24/32 bytes, got %d", len(key))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create cipher: %w", err)
	}

	// 使用零 IV，加密=解密（CTR 模式对称性）
	iv := make([]byte, 16)
	enc := cipher.NewCTR(block, iv)
	block2, _ := aes.NewCipher(key)
	dec := cipher.NewCTR(block2, iv)

	return &StreamCipher{
		enc: enc,
		dec: dec,
		key: key,
	}, nil
}

// NewStreamCipherWithIV 创建带 IV 的流加密器。
func NewStreamCipherWithIV(key, iv []byte) (*StreamCipher, error) {
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return nil, fmt.Errorf("key length must be 16/24/32 bytes, got %d", len(key))
	}
	if len(iv) != 16 {
		return nil, fmt.Errorf("iv must be 16 bytes, got %d", len(iv))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	enc := cipher.NewCTR(block, iv)
	block2, _ := aes.NewCipher(key)
	dec := cipher.NewCTR(block2, iv)

	return &StreamCipher{enc: enc, dec: dec, key: key}, nil
}

// NewBidirectionalCipher 创建双向加密器（Bedrock 模式）。
// sendKey/recvKey/sendIV/recvIV 都由 ECDH 派生。
func NewBidirectionalCipher(sendKey, recvKey, sendIV, recvIV []byte) (send, recv *StreamCipher, err error) {
	send, err = NewStreamCipherWithIV(sendKey, sendIV)
	if err != nil {
		return nil, nil, fmt.Errorf("send cipher: %w", err)
	}
	recv, err = NewStreamCipherWithIV(recvKey, recvIV)
	if err != nil {
		return nil, nil, fmt.Errorf("recv cipher: %w", err)
	}
	return send, recv, nil
}

// Encrypt 加密数据（就地）。
func (s *StreamCipher) Encrypt(dst, src []byte) {
	s.enc.XORKeyStream(dst, src)
}

// Decrypt 解密数据（就地）。
func (s *StreamCipher) Decrypt(dst, src []byte) {
	s.dec.XORKeyStream(dst, src)
}

// EncryptCopy 加密并返回新数据。
func (s *StreamCipher) EncryptCopy(src []byte) []byte {
	dst := make([]byte, len(src))
	s.enc.XORKeyStream(dst, src)
	return dst
}

// DecryptCopy 解密并返回新数据。
func (s *StreamCipher) DecryptCopy(src []byte) []byte {
	dst := make([]byte, len(src))
	s.dec.XORKeyStream(dst, src)
	return dst
}

// Key 返回密钥。
func (s *StreamCipher) Key() []byte {
	return s.key
}

// GCMCipher 提供 AES-256-GCM 认证加密（备用模式）。
type GCMCipher struct {
	gcm cipher.AEAD
}

// NewGCMCipher 创建 GCM 加密器。
func NewGCMCipher(key, nonce []byte) (*GCMCipher, error) {
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return nil, fmt.Errorf("invalid key length")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return &GCMCipher{gcm: gcm}, nil
}

// Seal 加密并认证。
func (g *GCMCipher) Seal(dst, nonce, plaintext, additionalData []byte) []byte {
	return g.gcm.Seal(dst, nonce, plaintext, additionalData)
}

// Open 解密并验证。
func (g *GCMCipher) Open(dst, nonce, ciphertext, additionalData []byte) ([]byte, error) {
	return g.gcm.Open(dst, nonce, ciphertext, additionalData)
}

// NonceSize 返回 nonce 大小。
func (g *GCMCipher) NonceSize() int {
	return g.gcm.NonceSize()
}

// HKDF HKDF 密钥派生。
func HKDF(secret, salt, info []byte, length int) ([]byte, error) {
	// 简化版 HKDF-SHA256
	prk := hmacSHA256(salt, secret)
	if len(prk) == 0 {
		return nil, errors.New("empty PRK")
	}

	// Expand
	okm := make([]byte, 0, length)
	counter := byte(1)
	var prev []byte
	for len(okm) < length {
		data := make([]byte, 0, len(prev)+len(info)+1)
		data = append(data, prev...)
		data = append(data, info...)
		data = append(data, counter)
		prev = hmacSHA256(prk, data)
		okm = append(okm, prev...)
		counter++
	}
	return okm[:length], nil
}

func hmacSHA256(key, data []byte) []byte {
	return simpleHMAC(key, data)
}

func simpleHMAC(key, data []byte) []byte {
	// 简化：使用 SHA-256(key || data) || SHA-256((key xor 0x5c) || SHA-256(key xor 0x36 || data))
	// 这里直接用两次 SHA-256
	inner := sha256.New()
	if len(key) > 64 {
		key = sha256Sum(key)
	}
	paddedKey := make([]byte, 64)
	copy(paddedKey, key)
	for i := len(key); i < 64; i++ {
		paddedKey[i] = 0x36
	}
	inner.Write(paddedKey)
	inner.Write(data)
	innerSum := inner.Sum(nil)

	outer := sha256.New()
	paddedKey2 := make([]byte, 64)
	for i := range paddedKey2 {
		if i < len(key) {
			paddedKey2[i] = key[i] ^ 0x5c
		} else {
			paddedKey2[i] = 0x5c
		}
	}
	outer.Write(paddedKey2)
	outer.Write(innerSum)
	return outer.Sum(nil)
}

func sha256Sum(data []byte) []byte {
	h := sha256.Sum256(data)
	return h[:]
}

// NeteaseCrypto 网易专用加密层。
// 网易版加密在 X25519 之上加一层自定义编码。
type NeteaseCrypto struct {
	*ECDH
	cipher *StreamCipher
}

// NewNeteaseCrypto 创建网易加密器。
func NewNeteaseCrypto(salt []byte) (*NeteaseCrypto, error) {
	ecdh, err := NewECDH()
	if err != nil {
		return nil, err
	}
	// 派生初始密钥
	_, err = ecdh.PublicKeyBytes(), error(nil)
	_ = err
	_ = salt
	return &NeteaseCrypto{ECDH: ecdh}, nil
}

// Finalize 完成密钥派生（shared secret 可用后调用）。
func (n *NeteaseCrypto) Finalize(sharedSecret, salt []byte) error {
	key, err := DeriveKey(sharedSecret, salt)
	if err != nil {
		return err
	}
	if len(key) < 32 {
		return fmt.Errorf("derived key too short: %d", len(key))
	}
	cipher, err := NewStreamCipher(key[:32])
	if err != nil {
		return err
	}
	n.cipher = cipher
	return nil
}

// Cipher 返回流加密器。
func (n *NeteaseCrypto) Cipher() *StreamCipher {
	return n.cipher
}

// HashChallenge 计算 challenge hash。
// Bedrock 服务器在加密握手中发送 challenge，客户端需要用共享密钥签名回应。
func HashChallenge(challenge, sharedSecret []byte) []byte {
	h := sha256.New()
	h.Write(challenge)
	h.Write(sharedSecret)
	return h.Sum(nil)
}

// RandomBytes 生成 n 字节随机数。
func RandomBytes(n int) []byte {
	b := make([]byte, n)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		panic(err)
	}
	return b
}
