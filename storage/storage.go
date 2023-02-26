package storage

import "context"

type Storage interface {
	GetLastTransactionLt(ctx context.Context) (uint64, error)
	CheckTransaction(ctx context.Context, txHash string) (bool, error)

	GetTransactionBySign(ctx context.Context, sign string) (*Tx, error)

	AckTransaction(ctx context.Context, txHash string) error
	StoreTransaction(ctx context.Context, tx Tx) error
}
