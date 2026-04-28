package secret

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/crypto/ssh"
)

func checkPrivateKeyProtection(keyPath string) error {
	data, err := os.ReadFile(keyPath)
	if err != nil {
		return fmt.Errorf("read private key: %w", err)
	}

	content := string(data)
	if strings.Contains(content, "ENCRYPTED") {
		return nil
	}

	_, err = ssh.ParseRawPrivateKey(data)
	if err == nil {
		fmt.Fprintf(os.Stderr, "\nWARNING: Your SSH private key is not password-protected.\n")
		fmt.Fprintf(os.Stderr, "  Key: %s\n", keyPath)
		fmt.Fprintf(os.Stderr, "  Anyone with access to this file can decrypt your secret files.\n")
		fmt.Fprintf(os.Stderr, "\n  To add password protection:\n")
		fmt.Fprintf(os.Stderr, "  1. Backup your key: cp %s %s.backup\n", keyPath, keyPath)
		fmt.Fprintf(os.Stderr, "  2. Add password: ssh-keygen -p -f %s\n\n", keyPath)
		return nil
	}

	if strings.Contains(content, "BEGIN OPENSSH PRIVATE KEY") ||
		strings.Contains(content, "BEGIN RSA PRIVATE KEY") ||
		strings.Contains(content, "BEGIN EC PRIVATE KEY") {
		return nil
	}

	return fmt.Errorf("unable to verify key protection status")
}
