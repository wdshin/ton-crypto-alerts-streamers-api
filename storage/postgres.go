package storage

import (
	"context"
	"database/sql"

	_ "github.com/lib/pq"
)

type Postgres struct {
	db *sql.DB
}

func NewPostgres(connStr string) (*Postgres, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	pg := &Postgres{
		db: db,
	}

	var _ Storage = pg

	return pg, nil
}

func (p *Postgres) GetLastTransactionLt(ctx context.Context) (uint64, error) {
	var lt uint64
	err := p.db.QueryRowContext(
		ctx,
		"SELECT lt FROM transactions ORDER BY lt DESC LIMIT 1",
	).Scan(&lt)
	if err != nil {
		return 0, err
	}

	return lt, nil
}

func (p *Postgres) GetTransactionBySign(ctx context.Context, sign string) (*Tx, error) {
	var tx Tx
	err := p.db.QueryRowContext(
		ctx,
		"SELECT tx_hash, sign, wallet_address, amount, lt, acked, created_at FROM transactions WHERE sign = $1",
		sign,
	).Scan(
		&tx.TxHash,
		&tx.Sign,
		&tx.WalletAddress,
		&tx.Amount,
		&tx.Lt,
		&tx.Acked,
		&tx.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &tx, nil
}

func (p *Postgres) CheckTransaction(ctx context.Context, txId string) (bool, error) {
	var txHash string
	err := p.db.QueryRowContext(
		ctx,
		"SELECT tx_hash FROM transactions WHERE tx_hash = $1",
		txId,
	).Scan(&txHash)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (p *Postgres) AckTransaction(ctx context.Context, txId string) error {
	_, err := p.db.ExecContext(
		ctx,
		"UPDATE transactions SET acked = true WHERE tx_hash = $1",
		txId,
	)
	if err != nil {
		return err
	}

	return nil
}

func (p *Postgres) StoreTransaction(ctx context.Context, tx Tx) error {
	_, err := p.db.ExecContext(
		ctx,
		`INSERT INTO transactions (
			tx_hash,
			sign,
			wallet_address,
			amount,
			lt
		) VALUES ($1, $2, $3, $4, $5)`,
		tx.TxHash,
		tx.Sign,
		tx.WalletAddress,
		tx.Amount,
		tx.Lt,
	)
	if err != nil {
		return err
	}

	return nil
}

func (p *Postgres) StoreStreamer(ctx context.Context, streamer Streamer) error {
	_, err := p.db.ExecContext(
		ctx,
		`INSERT INTO streamers (
			wallet_address,
			client_id
		) VALUES ($1, $2)`,
		streamer.WalletAddress,
		streamer.StreamerId,
	)
	if err != nil {
		return err
	}

	return nil
}

func (p *Postgres) GetStreamerClientIdByWalletAddress(ctx context.Context, walletAddress string) (string, error) {
	var clientId string

	err := p.db.QueryRowContext(
		ctx,
		"SELECT client_id FROM streamers WHERE wallet_address = $1",
		walletAddress,
	).Scan(
		&clientId,
	)
	if err != nil {
		return "", err
	}

	return clientId, nil
}
