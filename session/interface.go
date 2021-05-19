package session

import (
	"context"
)

type Session interface {
	Get(ctx context.Context, key string) (string, error)
	Put(ctx context.Context, key string, value string) error
	Delete(ctx context.Context, key string) error
	Init(ctx context.Context, setting map[string]interface{}) error
	Close(ctx context.Context) error
}
