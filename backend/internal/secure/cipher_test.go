package secure

import (
	"bytes"
	"encoding/base64"
	"testing"
)

func TestCipherRoundTrip(t *testing.T) {
	c, err := NewCipher(bytes.Repeat([]byte{7}, 32))
	if err != nil {
		t.Fatal(err)
	}
	enc, err := c.Encrypt("access-token-123")
	if err != nil {
		t.Fatal(err)
	}
	got, err := c.Decrypt(enc)
	if err != nil {
		t.Fatal(err)
	}
	if got != "access-token-123" {
		t.Fatalf("got %q", got)
	}
}

func TestCipherRejectsTampering(t *testing.T) {
	c, _ := NewCipher(bytes.Repeat([]byte{1}, 32))
	enc, _ := c.Encrypt("secret")
	raw, _ := base64.StdEncoding.DecodeString(enc)
	raw[len(raw)-1] ^= 0xff
	if _, err := c.Decrypt(base64.StdEncoding.EncodeToString(raw)); err == nil {
		t.Fatal("expected error for tampered ciphertext")
	}
}

func TestCipherRejectsWrongKeyLength(t *testing.T) {
	if _, err := NewCipher([]byte("short")); err == nil {
		t.Fatal("expected error")
	}
}
