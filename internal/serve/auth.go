package serve

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/eljojo/rememory/internal/core"
)

// knownPlaintext is the fixed string encrypted with the admin password.
// Verification works by decrypting admin.age and checking the result matches.
const knownPlaintext = "rememory-admin"

// SetPassword encrypts a known plaintext with the given password and saves it.
// Returns an error if a password is already set.
func SetPassword(store *Store, password string) error {
	if password == "" {
		return fmt.Errorf("password cannot be empty")
	}

	path := store.AdminFilePath()
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("admin password is already set")
	}

	var buf bytes.Buffer
	if err := core.Encrypt(&buf, strings.NewReader(knownPlaintext), password); err != nil {
		return fmt.Errorf("encrypting admin password: %w", err)
	}

	if err := os.WriteFile(path, buf.Bytes(), 0600); err != nil {
		return fmt.Errorf("writing admin password file: %w", err)
	}

	return nil
}

// CheckPassword verifies the given password against the stored admin password.
func CheckPassword(store *Store, password string) bool {
	if password == "" {
		return false
	}

	path := store.AdminFilePath()
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}

	decrypted, err := core.DecryptBytes(data, password)
	if err != nil {
		return false
	}

	return string(decrypted) == knownPlaintext
}

// IsSetup returns true if an admin password has been configured.
func IsSetup(store *Store) bool {
	_, err := os.Stat(store.AdminFilePath())
	return err == nil
}
