package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gabrandalisse/rant/internal/cli"
	"github.com/gabrandalisse/rant/internal/crypto"
	"github.com/gabrandalisse/rant/internal/keyring"
	"github.com/gabrandalisse/rant/internal/rant"
	"github.com/gabrandalisse/rant/internal/sqlite"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("resolve home directory: %w", err)
	}

	dbPath := filepath.Join(homeDir, ".local", "share", "rant", "vault.db")

	keyStore := keyring.NewOSKeyStore()
	masterKey, err := keyStore.GetOrCreateMasterKey()
	if err != nil {
		return err
	}

	cipher, err := crypto.NewAESGCMCipher(masterKey)
	if err != nil {
		return err
	}

	store, err := sqlite.Open(dbPath)
	if err != nil {
		return err
	}
	defer func() { _ = store.Close() }()

	svc := rant.NewService(store, cipher, keyStore)
	return cli.Execute(svc)
}
