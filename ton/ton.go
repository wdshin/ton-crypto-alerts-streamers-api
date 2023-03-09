package ton

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/vladtenlive/ton-donate/storage"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/ton"
)

type Connector struct {
	Address      *address.Address
	Network      string
	Client       *ton.APIClient
	storage      storage.Storage
	mongoStorage *storage.MongoStorage
}

func New(
	ctx context.Context,
	watchAddress string,
	storage storage.Storage,
	mongoStorage *storage.MongoStorage,
) (*Connector, error) {
	connPool := liteclient.NewConnectionPool()

	configUrl := "https://ton-blockchain.github.io/global.config.json"
	err := connPool.AddConnectionsFromConfigUrl(ctx, configUrl)
	if err != nil {
		return nil, err
	}

	client := ton.NewAPIClient(connPool)

	return &Connector{
		storage:      storage,
		mongoStorage: mongoStorage,
		Address:      address.MustParseAddr(watchAddress),
		Client:       client,
		Network:      "mainnet",
	}, nil
}

func (c *Connector) Start(ctx context.Context, d time.Duration) {
	for {
		go c.GetTransactions(ctx)
		time.Sleep(d)
	}
}

func (c *Connector) GetTransactions(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in f", r)
		}
	}()

	block, err := c.Client.CurrentMasterchainInfo(ctx)
	if err != nil {
		log.Println(err)
		return
	}

	account, err := c.Client.GetAccount(ctx, block, c.Address)
	if err != nil {
		log.Println(err)
		return
	}

	if account == nil {
		return
	}

	hash := account.LastTxHash
	lt := account.LastTxLT

	seenLt, err := c.storage.GetLastTransactionLt(ctx)
	if err != nil {
		log.Println(err)
	}

	if seenLt >= lt {
		return
	}

	txs, err := c.Client.ListTransactions(ctx, c.Address, 100, lt, hash)
	if err != nil {
		log.Println(err)
		return
	}

	for _, tx := range txs {

		if len(tx.IO.Out) > 0 {
			continue
		}

		txHash := fmt.Sprintf("%x", tx.Hash)
		if ok, err := c.storage.CheckTransaction(ctx, txHash); err == nil && ok {
			continue
		}

		txInfo := tx.IO.In.AsInternal()

		data := strings.Split(txInfo.Comment(), ":")
		if len(data) != 2 {
			continue
		}

		streamerAddress := data[0]
		messageSign := data[1]

		transaction := storage.Tx{
			Sign:          messageSign,
			TxHash:        txHash,
			WalletAddress: streamerAddress,
			Amount:        txInfo.Amount.NanoTON().Uint64(),
			Lt:            tx.LT,
			Acked:         false,
		}

		err = c.storage.StoreTransaction(ctx, transaction)
		if err != nil {
			log.Println(err)
			continue
		}
	}
}
