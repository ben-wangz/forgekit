package secret

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha512"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/crypto/curve25519"
	"golang.org/x/crypto/nacl/box"
	"golang.org/x/crypto/ssh"
)

const (
	magicHeader = "SECRET-V1"
	nonceSize   = 24
)

func encryptFile(filePath string) error {
	if !strings.Contains(filePath, ".secret.") {
		return fmt.Errorf("file must match *.secret.* pattern")
	}
	if strings.HasSuffix(filePath, ".enc") {
		return fmt.Errorf("file is already encrypted")
	}

	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	if err := checkPrivateKeyProtection(cfg.PrivateKeyPath); err != nil {
		return err
	}

	recipientPub, err := loadPublicKey(cfg.PublicKeyPath)
	if err != nil {
		return err
	}

	plaintext, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	ephemeralPub, ephemeralPriv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return fmt.Errorf("generate ephemeral key: %w", err)
	}

	var nonce [nonceSize]byte
	if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
		return fmt.Errorf("generate nonce: %w", err)
	}

	encrypted := box.Seal(nil, plaintext, &nonce, recipientPub, ephemeralPriv)
	outPath := filePath + ".enc"

	outFile, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("create output: %w", err)
	}
	defer outFile.Close()

	if _, err := outFile.WriteString(magicHeader + "\n"); err != nil {
		return fmt.Errorf("write header: %w", err)
	}
	if _, err := outFile.Write(ephemeralPub[:]); err != nil {
		return fmt.Errorf("write ephemeral key: %w", err)
	}
	if _, err := outFile.Write(nonce[:]); err != nil {
		return fmt.Errorf("write nonce: %w", err)
	}

	dataLen := uint64(len(encrypted))
	if err := binary.Write(outFile, binary.LittleEndian, dataLen); err != nil {
		return fmt.Errorf("write data length: %w", err)
	}
	if _, err := outFile.Write(encrypted); err != nil {
		return fmt.Errorf("write encrypted data: %w", err)
	}

	fmt.Printf("Encrypted: %s -> %s\n", filePath, outPath)
	return nil
}

func decryptFile(filePath string) error {
	if !strings.HasSuffix(filePath, ".enc") {
		return fmt.Errorf("file must end with .enc")
	}

	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	if err := checkPrivateKeyProtection(cfg.PrivateKeyPath); err != nil {
		return err
	}

	recipientPriv, err := loadPrivateKey(cfg.PrivateKeyPath)
	if err != nil {
		return err
	}

	inFile, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer inFile.Close()

	header := make([]byte, len(magicHeader)+1)
	if _, err := io.ReadFull(inFile, header); err != nil {
		return fmt.Errorf("read header: %w", err)
	}
	if string(header[:len(magicHeader)]) != magicHeader {
		return fmt.Errorf("invalid file format")
	}

	var ephemeralPub [32]byte
	if _, err := io.ReadFull(inFile, ephemeralPub[:]); err != nil {
		return fmt.Errorf("read ephemeral key: %w", err)
	}

	var nonce [nonceSize]byte
	if _, err := io.ReadFull(inFile, nonce[:]); err != nil {
		return fmt.Errorf("read nonce: %w", err)
	}

	var dataLen uint64
	if err := binary.Read(inFile, binary.LittleEndian, &dataLen); err != nil {
		return fmt.Errorf("read data length: %w", err)
	}

	encrypted := make([]byte, dataLen)
	if _, err := io.ReadFull(inFile, encrypted); err != nil {
		return fmt.Errorf("read encrypted data: %w", err)
	}

	plaintext, ok := box.Open(nil, encrypted, &nonce, &ephemeralPub, recipientPriv)
	if !ok {
		return fmt.Errorf("decryption failed")
	}

	outPath := strings.TrimSuffix(filePath, ".enc")
	if err := os.WriteFile(outPath, plaintext, 0o600); err != nil {
		return fmt.Errorf("write output: %w", err)
	}

	fmt.Printf("Decrypted: %s -> %s\n", filePath, outPath)
	return nil
}

func loadPublicKey(path string) (*[32]byte, error) {
	privPath := strings.TrimSuffix(path, ".pub")
	data, err := os.ReadFile(privPath)
	if err != nil {
		return nil, fmt.Errorf("read private key: %w", err)
	}

	privKey, err := ssh.ParseRawPrivateKey(data)
	if err != nil {
		return nil, fmt.Errorf("parse private key: %w", err)
	}

	ed25519Priv, ok := privKey.(*ed25519.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("only ed25519 keys are supported")
	}

	h := sha512.Sum512((*ed25519Priv)[:32])
	h[0] &= 248
	h[31] &= 127
	h[31] |= 64

	var curve25519Priv [32]byte
	copy(curve25519Priv[:], h[:32])

	var curve25519Pub [32]byte
	curve25519.ScalarBaseMult(&curve25519Pub, &curve25519Priv)

	return &curve25519Pub, nil
}

func loadPrivateKey(path string) (*[32]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read private key: %w", err)
	}

	privKey, err := ssh.ParseRawPrivateKey(data)
	if err != nil {
		return nil, fmt.Errorf("parse private key: %w", err)
	}

	ed25519Priv, ok := privKey.(*ed25519.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("only ed25519 keys are supported")
	}

	h := sha512.Sum512((*ed25519Priv)[:32])
	h[0] &= 248
	h[31] &= 127
	h[31] |= 64

	var curve25519Priv [32]byte
	copy(curve25519Priv[:], h[:32])
	return &curve25519Priv, nil
}
