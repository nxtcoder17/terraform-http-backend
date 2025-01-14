package store

import (
	"context"
	"net/http"
)

type Metadata struct {
	Dir string
}

func MetadataFromRequest(req *http.Request) *Metadata {
	return &Metadata{
		Dir: req.URL.Query().Get("dir"),
	}
}

type LockStateArgs struct {
	Lockfile string
	Body     []byte
}

type UnlockStateArgs struct {
	Lockfile string
}

type Store interface {
	LockState(ctx context.Context, args LockStateArgs) (*Lock, error)
	UnlockState(ctx context.Context, args UnlockStateArgs) error

	ReadState(name string) ([]byte, error)
	WriteState(name string, b []byte) error
	DeleteState(name string) error
}

// Lock type generated after logging LOCK request body
type Lock struct {
	ID        string
	Operation string
	Info      string
	Who       string
	Version   string
	Created   string
	Path      string
}
