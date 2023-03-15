package ton

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
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
	notifier     *Notifier
}

func New(
	ctx context.Context,
	watchAddress string,
	storage storage.Storage,
	mongoStorage *storage.MongoStorage,
	notifier *Notifier,
) (*Connector, error) {
	connPool := liteclient.NewConnectionPool()
	configUrl := os.Getenv("TON_CONFIG_URL")

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
		notifier:     notifier,
		Network:      os.Getenv("TON_NET"),
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
	// return

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

	txs, err := c.Client.ListTransactions(ctx, c.Address, 100, lt, hash)
	if err != nil {
		log.Println(err)
		return
	}

	for _, tx := range txs {
		transaction := parseBody(tx)

		donation, err := c.mongoStorage.GetDonationBySign(ctx, transaction.Sign)
		if err != nil {
			log.Println("(1) GetDonationBySign: ", err)
			// ToDo: just skip for now, later we can figure out
			continue
		}

		if (donation != nil && donation.Sign == "") || transaction.Sign == "" { // ToDO: why some transactions have empty sign and wallet?
			continue
		}

		streamerId, err := getOrLoadStreamerId(ctx, c, donation, transaction)
		if err != nil {
			log.Println("Skip transaction processing when wallet address is empty")
			continue
		}

		_, err = c.mongoStorage.SaveDonation(ctx, transaction, streamerId)
		if err != nil {
			log.Println("Failed to save donation transaction info: ", err)
		}

		donation, err = c.mongoStorage.GetDonationBySign(ctx, transaction.Sign)
		if err != nil {
			log.Println("(2) GetDonationBySign: ", err)
			// ToDo: just skip for now, later we can figure out
			continue
		}

		if donation.Acked {
			// No need to process acked transaction
			continue
		}

		// fmt.Println("transaction: ", transaction)
		donationAmount := uint64(transaction.Amount)
		notificationReq := NotificationRequest{
			Id:         fmt.Sprintf(transaction.TxHash), // or could be d.Sign depends on Storage
			Amount:     donationAmount,
			Text:       transaction.Message,
			Nickname:   donation.From,
			StreamerId: donation.StreamerId,
		}
		err = c.notifier.Send(notificationReq)
		if err != nil {
			var notificationError NotificationError
			if errors.As(err, &notificationError) {
				log.Println("Resubmit request id: ", notificationError.Id)
			} else {
				log.Println(err)
			}
		} else {
			_, err = c.mongoStorage.AckDonation(ctx, transaction)
			if err != nil {
				log.Println("Failed to ack donation with sign: ", transaction.Sign)
				return
			}

			_, err = c.mongoStorage.AddToCurrentAmount(ctx, donation.StreamerId, donationAmount)
			if err != nil {
				log.Println("Failed to add donation to widget total sum: ", transaction.Sign)
			}
		}
	}
}

func parseBody(trx *tlb.Transaction) storage.Tx {
	txHash := fmt.Sprintf("%x", trx.Hash)

	payload := trx.IO.In.Msg.Payload().BeginParse()
	payload.LoadUInt(32) // skip op code
	streamerAddress, err := payload.LoadAddr()
	if err == nil {
		// fmt.Println(streamerAddress)
	}
	sign, err := payload.LoadStringSnake()
	if err == nil {
		// fmt.Println(sign)
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

func getOrLoadStreamerId(ctx context.Context, c *Connector, donation *storage.Donation, transaction storage.Tx) (string, error) {
	if donation == nil || donation.StreamerId == "" {
		log.Println("Mapping streamer id by transaction wallet address. Possibly donation request failed to save.")
		streamer, err := c.mongoStorage.GetStreamerByWalletAddress(ctx, transaction.WalletAddress)
		if err != nil {
			log.Println("Failed to map streamer id by transaction wallet address.")
			return "", err
		}

		return streamer.StreamerId, nil
	}

	return donation.StreamerId, nil
}
