package storage

// import (
// 	"context"
// 	"time"

// 	"github.com/redis/go-redis/v9"
// )

// type Redis struct {
// 	client *redis.Client
// }

// func NewRedis(opts *redis.Options) *Redis {
// 	return &Redis{
// 		client: redis.NewClient(opts),
// 	}
// }

// func (r *Redis) GetTx(ctx context.Context, txHash string) (*Tx, error) {
// 	lt, err := r.client.Get(ctx, txHash).Result()
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &Tx{
// 		TxHash: txHash,
// 		Lt:     lt,
// 	}, nil
// }

// func (r *Redis) StoreTx(ctx context.Context, tx Tx) error {
// 	err := r.client.Set(
// 		ctx,
// 		tx.TxHash,
// 		tx.Lt,
// 		1*time.Hour,
// 	).Err()
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// func (r *Redis) GetNotificationId(ctx context.Context, walletAddress string) (string, error) {
// 	notificationId, err := r.client.Get(ctx, walletAddress).Result()
// 	if err != nil {
// 		return "", err
// 	}

// 	return notificationId, nil
// }

// func (r *Redis) StoreNotificationId(ctx context.Context, streamer Streamer) error {
// 	err := r.client.Set(
// 		ctx,
// 		streamer.WalletAddress,
// 		streamer.NotificationId,
// 		1*time.Hour,
// 	).Err()

// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }
