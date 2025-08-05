package storage

import (
	"context"
	"time"
)

// StorageStrategy define a interface para diferentes implementações de storage
type StorageStrategy interface {
	// Increment incrementa o contador de requisições para uma chave específica
	Increment(ctx context.Context, key string, expiration time.Duration) (int64, error)

	// Get retorna o valor atual do contador para uma chave específica
	Get(ctx context.Context, key string) (int64, error)

	// Set define um valor para uma chave com expiração
	Set(ctx context.Context, key string, value int64, expiration time.Duration) error

	// Exists verifica se uma chave existe
	Exists(ctx context.Context, key string) (bool, error)

	// Delete remove uma chave
	Delete(ctx context.Context, key string) error

	// Close fecha a conexão com o storage
	Close() error
}
