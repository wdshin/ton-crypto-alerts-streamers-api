package storage

import "time"

type Tx struct {
	Sign          string
	TxHash        string
	WalletAddress string
	Amount        uint64
	Lt            uint64
	Acked         bool
	CreatedAt     time.Time
}
