package ton

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/vladtenlive/ton-donate/storage"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/tlb"
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

	// configUrl := "https://ton-blockchain.github.io/global.config.json"
	configUrl := "https://ton-blockchain.github.io/testnet-global.config.json"

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
		// Network:      "mainnet",
		Network: "testnet",
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

	txs, err := c.Client.ListTransactions(ctx, c.Address, 5, lt, hash)
	if err != nil {
		log.Println(err)
		return
	}

	for _, tx := range txs {
		transaction := parseBody(tx)

		fmt.Println("transaction: ", transaction)

		_, err := c.mongoStorage.SaveDonation(ctx, transaction)
		if err != nil {
			log.Println(err)
		}

		// ToDo: Publish notify event
	}
}

func parseBody(trx *tlb.Transaction) storage.Tx {
	txHash := fmt.Sprintf("%x", trx.Hash)

	payload := trx.IO.In.Msg.Payload().BeginParse()
	payload.LoadUInt(32) // skip op code
	streamerAddress, err := payload.LoadAddr()
	if err == nil {
		fmt.Println(streamerAddress)
	}
	sign, err := payload.LoadStringSnake()
	if err == nil {
		fmt.Println(sign)
	}

	txInfo := trx.IO.In.AsInternal()

	transaction := storage.Tx{
		Sign:          sign,
		Message:       txInfo.Comment(),
		TxHash:        txHash,
		WalletAddress: fmt.Sprintf("%s", streamerAddress),
		Amount:        txInfo.Amount.NanoTON().Uint64(),
		Lt:            trx.LT,
		Acked:         false,
	}

	return transaction
}
