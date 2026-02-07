package crypt

import (
	"bytes"
	"encoding/base64"
	"os"
	"testing"
)

func TestManager_EncryptDecrypt_RoundTrip(t *testing.T) {
	key := bytes.Repeat([]byte{1}, KeySize)

	m, err := NewManager(key)
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}

	plain := []byte("hello")
	aad := []byte("aad")

	nonce, ct, err := m.Encrypt(plain, aad)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	if len(nonce) != NonceSize {
		t.Fatalf("nonce len=%d want %d", len(nonce), NonceSize)
	}

	got, err := m.Decrypt(nonce, ct, aad)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}
	if !bytes.Equal(got, plain) {
		t.Fatalf("plain mismatch: got=%q want=%q", got, plain)
	}
}

func TestManager_Decrypt_WrongAAD_Fails(t *testing.T) {
	key := bytes.Repeat([]byte{2}, KeySize)
	m, _ := NewManager(key)

	nonce, ct, _ := m.Encrypt([]byte("hello"), []byte("aad"))
	_, err := m.Decrypt(nonce, ct, []byte("wrong"))
	if err == nil {
		t.Fatalf("expected error on wrong aad")
	}
}

func TestLoadAESKeyFromFile_Raw32(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/key"

	raw := bytes.Repeat([]byte{'a'}, KeySize)
	if err := writeFile(path, raw); err != nil {
		t.Fatalf("write: %v", err)
	}

	k, err := LoadAESKeyFromFile(path)
	if err != nil {
		t.Fatalf("LoadAESKeyFromFile: %v", err)
	}
	if !bytes.Equal(k, raw) {
		t.Fatalf("key mismatch")
	}
}

func TestLoadAESKeyFromFile_Base64(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/key"

	raw := bytes.Repeat([]byte{7}, KeySize)
	b64 := []byte(base64.StdEncoding.EncodeToString(raw))
	if err := writeFile(path, b64); err != nil {
		t.Fatalf("write: %v", err)
	}

	k, err := LoadAESKeyFromFile(path)
	if err != nil {
		t.Fatalf("LoadAESKeyFromFile: %v", err)
	}
	if !bytes.Equal(k, raw) {
		t.Fatalf("key mismatch")
	}
}

func writeFile(path string, b []byte) error {
	return osWriteFile(path, b, 0o600)
}

var osWriteFile = func(path string, data []byte, perm uint32) error {
	return os.WriteFile(path, data, os.FileMode(perm))
}
