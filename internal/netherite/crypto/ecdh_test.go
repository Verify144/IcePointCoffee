package crypto

import (
	"bytes"
	"testing"
)

func TestECDH(t *testing.T) {
	alice, err := NewECDH()
	if err != nil {
		t.Fatal(err)
	}
	bob, err := NewECDH()
	if err != nil {
		t.Fatal(err)
	}

	// Alice 用 Bob 的公钥派生共享密钥
	secretA, err := alice.ECDH(bob.PublicKeyBytes())
	if err != nil {
		t.Fatal(err)
	}

	// Bob 用 Alice 的公钥派生共享密钥
	secretB, err := bob.ECDH(alice.PublicKeyBytes())
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(secretA, secretB) {
		t.Fatal("shared secrets don't match")
	}
}

func TestStreamCipher(t *testing.T) {
	key := RandomBytes(32)
	cipher, err := NewStreamCipher(key)
	if err != nil {
		t.Fatal(err)
	}

	plaintext := []byte("Hello, Bedrock! 这是一个测试消息。")
	encrypted := cipher.EncryptCopy(plaintext)
	if bytes.Equal(encrypted, plaintext) {
		t.Fatal("encrypted equals plaintext")
	}

	decrypted := cipher.DecryptCopy(encrypted)
	if !bytes.Equal(decrypted, plaintext) {
		t.Fatalf("decrypted != plaintext: got %x, want %x", decrypted, plaintext)
	}
}

func TestDeriveKey(t *testing.T) {
	secret := RandomBytes(32)
	salt := RandomBytes(16)
	key, err := DeriveKey(secret, salt)
	if err != nil {
		t.Fatal(err)
	}
	if len(key) < 16 {
		t.Errorf("key too short: %d", len(key))
	}
}

func TestRandomBytes(t *testing.T) {
	b := RandomBytes(32)
	if len(b) != 32 {
		t.Errorf("RandomBytes(32) returned %d bytes", len(b))
	}
	// 每个调用应该不同
	b2 := RandomBytes(32)
	if bytes.Equal(b, b2) {
		t.Error("RandomBytes returned same value twice")
	}
}

func TestHKDF(t *testing.T) {
	secret := RandomBytes(32)
	salt := RandomBytes(16)
	info := []byte("BedrockRaknet")

	key1, err := HKDF(secret, salt, info, 32)
	if err != nil {
		t.Fatal(err)
	}
	if len(key1) != 32 {
		t.Errorf("HKDF returned %d bytes, want 32", len(key1))
	}

	// 同样输入同样输出
	key2, err := HKDF(secret, salt, info, 32)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(key1, key2) {
		t.Error("HKDF not deterministic")
	}

	// 不同 salt 不同输出
	key3, err := HKDF(secret, RandomBytes(16), info, 32)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(key1, key3) {
		t.Error("Different salt should give different key")
	}
}

func TestHashChallenge(t *testing.T) {
	challenge := []byte("server-challenge-123")
	secret := RandomBytes(32)
	h := HashChallenge(challenge, secret)
	if len(h) != 32 {
		t.Errorf("HashChallenge returned %d bytes, want 32", len(h))
	}
}
