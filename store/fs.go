package store

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/nxtcoder17/terraform-backend-http/pkg/encryption"
)

type FileSystemStore struct {
	encryption.Cipher
}

func NewFileSystemStore(encryptionKey []byte) (*FileSystemStore, error) {
	c, err := encryption.NewAESCipher(encryptionKey)
	if err != nil {
		return nil, err
	}

	return &FileSystemStore{Cipher: c}, nil
}

func fileExists(p string) bool {
	fi, err := os.Stat(p)
	if err != nil {
		return false
	}
	return !fi.IsDir()
}

// LockState implements Store.
func (fs *FileSystemStore) LockState(ctx context.Context, args LockStateArgs) (*Lock, error) {
	if fileExists(args.Lockfile) {
		return nil, ErrAlreadyLocked
	}

	_, err := os.ReadDir(filepath.Dir(args.Lockfile))
	if err != nil {
		return nil, err
	}

	f, err := os.Create(args.Lockfile)
	if err != nil {
		return nil, errors.Join(ErrCreatingLockfile, err)
	}

	_, err = f.Write(args.Body)
	if err != nil {
		return nil, errors.Join(ErrWritingLockfile, err)
	}

	var lock Lock
	json.Unmarshal(args.Body, &lock)
	return &lock, nil
}

// UnlockState implements Store.
func (fs *FileSystemStore) UnlockState(ctx context.Context, args UnlockStateArgs) error {
	if !fileExists(args.Lockfile) {
		return ErrAlreadyUnlocked
	}

	return os.Remove(args.Lockfile)
}

// ReadState implements Store
func (f *FileSystemStore) ReadState(name string) ([]byte, error) {
	b, err := os.ReadFile(name)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	decrypted, err := f.Decrypt(b)
	if err != nil {
		return nil, err
	}

	return decrypted, nil
}

// WriteState implements Store.
func (fs *FileSystemStore) WriteState(name string, b []byte) error {
	encrypted, err := fs.Encrypt(b)
	if err != nil {
		return err
	}
	return os.WriteFile(name, encrypted, 0o666)
}

// DeleteState implements Store.
func (fs *FileSystemStore) DeleteState(name string) error {
	return os.Remove(name)
}

var _ Store = (*FileSystemStore)(nil)
