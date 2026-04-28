package secret

import "fmt"

func Run(args []string) error {
	if len(args) < 1 {
		printUsage()
		return nil
	}

	if args[0] == "help" || args[0] == "-h" || args[0] == "--help" {
		printUsage()
		return nil
	}

	cmd := args[0]
	subArgs := args[1:]

	switch cmd {
	case "encrypt":
		if len(subArgs) != 1 {
			return fmt.Errorf("usage: forgekit secret encrypt <file>")
		}
		return encryptFile(subArgs[0])
	case "decrypt":
		if len(subArgs) != 1 {
			return fmt.Errorf("usage: forgekit secret decrypt <file.enc>")
		}
		return decryptFile(subArgs[0])
	default:
		printUsage()
		return fmt.Errorf("unknown secret command: %s", cmd)
	}
}

func printUsage() {
	fmt.Println("Usage: forgekit secret <command> [args]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  encrypt <file>       Encrypt a single *.secret.* file")
	fmt.Println("  decrypt <file.enc>   Decrypt a single encrypted file")
	fmt.Println("  help                 Print help")
	fmt.Println()
	fmt.Println("Notes:")
	fmt.Println("  - Plaintext files should match *.secret.*")
	fmt.Println("  - Encrypted files should end with .enc")
	fmt.Println("  - Configure key path via SECRET_PRIVATE_KEY and SECRET_PUBLIC_KEY")
}
