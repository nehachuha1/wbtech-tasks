package postgres

import "context"

type IPostgresDatabase interface {
	CreateOrder(ctx context.Context, out chan interface{}, data []byte)
	GetOrder(ctx context.Context, out chan interface{}, data []byte)
	GrepOrdersToCache(ctx context.Context, out chan interface{})
}
