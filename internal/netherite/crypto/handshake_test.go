package crypto

import (
	"bytes"
	"testing"
)

func TestEncryptedHandshakeEncodeDecode(t *testing.T) {
	hs := &EncryptedHandshake{
		Cookie:    []byte{0x01, 0x02, 0x03, 0x04},
		PublicKey: bytes.Repeat([]byte{0xAA}, 32),
		Challenge: bytes.Repeat([]byte{0xBB}, 16),
	}

	encoded := hs.Encode()
	if len(encoded) != 53 {
		t.Errorf("encoded len: got %d want 53", len(encoded))
	}
	if encoded[0] != PacketIDConnectionRequestEncrypted {
		t.Errorf("packet ID mismatch")
	}

	decoded, err := DecodeEncryptedHandshake(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(decoded.Cookie, hs.Cookie) {
		t.Error("cookie mismatch")
	}
	if !bytes.Equal(decoded.PublicKey, hs.PublicKey) {
		t.Error("public key mismatch")
	}
	if !bytes.Equal(decoded.Challenge, hs.Challenge) {
		t.Error("challenge mismatch")
	}
}

func TestServerHandshakeEncodeDecode(t *testing.T) {
	hs := &ServerHandshake{
		ServerIdentity: []byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88},
		Cookie:         []byte{0x99, 0xAA, 0xBB, 0xCC},
		PublicKey:      bytes.Repeat([]byte{0xDD}, 32),
	}

	encoded := hs.Encode()
	if len(encoded) != 45 {
		t.Errorf("encoded len: got %d want 45", len(encoded))
	}

	decoded, err := DecodeServerHandshake(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(decoded.ServerIdentity, hs.ServerIdentity) {
		t.Error("server identity mismatch")
	}
	if !bytes.Equal(decoded.Cookie, hs.Cookie) {
		t.Error("cookie mismatch")
	}
	if !bytes.Equal(decoded.PublicKey, hs.PublicKey) {
		t.Error("public key mismatch")
	}
}

func TestNewConnectionAccepted(t *testing.T) {
	n := &NewConnectionAccepted{
		ServerIP:     [4]byte{127, 0, 0, 1},
		ServerPort:   19132,
		SendPingTime: 12345,
		SendPongTime: 67890,
	}
	encoded := n.Encode()
	if len(encoded) == 0 {
		t.Error("encoded empty")
	}
}

func TestBedrockKeyDerive(t *testing.T) {
	shared := RandomBytes(32)
	key := BedrockKeyDerive(shared)
	if len(key) != 32 {
		t.Errorf("key length: got %d want 32", len(key))
	}
	// 确定性
	key2 := BedrockKeyDerive(shared)
	if !bytes.Equal(key, key2) {
		t.Error("key derivation not deterministic")
	}
}

func TestNeteaseKeyDerive(t *testing.T) {
	shared := RandomBytes(32)
	serverChallenge := RandomBytes(16)
	clientChallenge := RandomBytes(16)

	key := NeteaseKeyDerive(shared, serverChallenge, clientChallenge)
	if len(key) != 32 {
		t.Errorf("key length: got %d want 32", len(key))
	}

	// 同样输入同样输出
	key2 := NeteaseKeyDerive(shared, serverChallenge, clientChallenge)
	if !bytes.Equal(key, key2) {
		t.Error("netease key derivation not deterministic")
	}

	// 不同 challenge 不同输出
	key3 := NeteaseKeyDerive(shared, RandomBytes(16), clientChallenge)
	if bytes.Equal(key, key3) {
		t.Error("Different challenge should give different key")
	}
}

func TestEndToEndEncryption(t *testing.T) {
	// 模拟完整 ECDH + 加密流程
	alice, _ := NewECDH()
	bob, _ := NewECDH()

	// 双方交换公钥，派生共享密钥
	aliceShared, _ := alice.ECDH(bob.PublicKeyBytes())
	bobShared, _ := bob.ECDH(alice.PublicKeyBytes())

	if !bytes.Equal(aliceShared, bobShared) {
		t.Fatal("shared secrets differ")
	}

	// 派生加密密钥
	aliceKey := BedrockKeyDerive(aliceShared)
	bobKey := BedrockKeyDerive(bobShared)

	if !bytes.Equal(aliceKey, bobKey) {
		t.Fatal("derived keys differ")
	}

	// Alice 加密消息
	aliceCipher, _ := NewStreamCipher(aliceKey)
	message := []byte("Hello from Alice to Bob!")
	encrypted := aliceCipher.EncryptCopy(message)

	// Bob 解密
	bobCipher, _ := NewStreamCipher(bobKey)
	decrypted := bobCipher.DecryptCopy(encrypted)

	if !bytes.Equal(message, decrypted) {
		t.Errorf("decryption failed: got %s, want %s", decrypted, message)
	}
}

func TestGetCookieKey(t *testing.T) {
	key1 := GetCookieKey("192.168.1.1:19132")
	key2 := GetCookieKey("192.168.1.1:19132")
	if !bytes.Equal(key1, key2) {
		t.Error("cookie key not deterministic for same IP")
	}

	key3 := GetCookieKey("192.168.1.2:19132")
	if bytes.Equal(key1, key3) {
		t.Error("Different IP should give different cookie key")
	}
}
