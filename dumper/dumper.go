package dumper

import (
	"context"
	"math/big"
	"strings"
	"time"

	"github.com/gridprotocol/dumper/database"
	"github.com/gridprotocol/dumper/logs"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

var (
	// blockNumber = big.NewInt(0)
	logger = logs.Logger("dumper")
)

type Dumper struct {
	endpoint        string
	contractABI     []abi.ABI
	contractAddress []common.Address
	// store           MapStore

	fromBlock *big.Int

	eventNameMap map[common.Hash]string
	indexedMap   map[common.Hash]abi.Arguments
}

// init a dumper with chain selected: local/dev
func NewGRIDDumper(chain_ep string, registerAddress, marketAddress common.Address) (dumper *Dumper, err error) {
	dumper = &Dumper{
		// store:        store,
		endpoint:     chain_ep,
		eventNameMap: make(map[common.Hash]string),
		indexedMap:   make(map[common.Hash]abi.Arguments),
	}

	// set contract
	dumper.contractAddress = []common.Address{registerAddress, marketAddress}

	// set abi
	registerABI, err := abi.JSON(strings.NewReader(RegisterABI))
	if err != nil {
		return dumper, err
	}

	marketABI, err := abi.JSON(strings.NewReader(MarketABI))
	if err != nil {
		return dumper, err
	}

	// set contract abi
	dumper.contractABI = []abi.ABI{registerABI, marketABI}

	// parse and save topics for each events
	for _, ABI := range dumper.contractABI {
		// each event
		for name, event := range ABI.Events {
			// save event in dumper
			dumper.eventNameMap[event.ID] = name
			var indexed abi.Arguments
			// each topic
			for _, arg := range ABI.Events[name].Inputs {
				if arg.Indexed {
					indexed = append(indexed, arg)
				}
			}
			// save topics for each event in dumper
			dumper.indexedMap[event.ID] = indexed
		}
	}

	// get block number from db
	logger.Debug("getting block number from db")
	blockNumber, err := database.GetBlockNumber()
	if err != nil {
		blockNumber = 0
	}
	logger.Debug("block number: ", blockNumber)

	// set block number for dumper
	dumper.fromBlock = big.NewInt(blockNumber)

	return dumper, nil
}

// sync db with block chain every 10 sec
func (d *Dumper) SubscribeGRID(ctx context.Context) {
	for {
		d.DumpGRID()

		select {
		case <-ctx.Done():
			return
		case <-time.After(10 * time.Second):
		}
	}
}

// dump all events of blocks into db
func (d *Dumper) DumpGRID() error {
	// dial chain
	client, err := ethclient.DialContext(context.TODO(), d.endpoint)
	if err != nil {
		logger.Debug(err.Error())
		return err
	}
	defer client.Close()

	// get current chain block number
	chainBlock, err := client.BlockNumber(context.Background())
	if err != nil {
		return err
	}

	// if no new chain block, return
	if d.fromBlock.Cmp(new(big.Int).SetUint64(chainBlock)) > 0 {
		logger.Debug("no new chain block, waiting..")
		return nil
	}

	logger.Debug("dump from block: ", d.fromBlock)

	// filter event logs from block
	events, err := client.FilterLogs(context.TODO(), ethereum.FilterQuery{
		FromBlock: d.fromBlock,
		Addresses: d.contractAddress,
	})
	if err != nil {
		logger.Debug(err.Error())
		return err
	}

	// record block
	lastBlock := d.fromBlock

	// parse each event
	for _, event := range events {
		// topic0 is the event name
		eventName, ok1 := d.eventNameMap[event.Topics[0]]
		if !ok1 {
			continue
		}

		switch eventName {
		case "Register":
			logger.Debug("==== Handle Register Event")
			err = d.HandleRegister(event)
			if err != nil {
				logger.Debug("handle register error: ", err.Error())
			}
		case "AddNode":
			logger.Debug("==== Handle Add Node Event")
			err = d.HandleAddNode(event)
			if err != nil {
				logger.Debug("handle addNode error: ", err.Error())
			}
		case "CreateOrder":
			logger.Debug("==== Handle Create Order Event")
			tx, _, err := client.TransactionByHash(context.TODO(), event.TxHash)
			if err != nil {
				logger.Debug("handle create order error: ", err.Error())
			}

			// get user address
			address, err := types.LatestSignerForChainID(tx.ChainId()).Sender(tx)
			if err != nil {
				logger.Debug(err.Error())
				return err
			}

			// store order info
			err = d.HandleCreateOrder(event, address)
			if err != nil {
				logger.Debug(err.Error())
			}
		default:
			continue
		}

		// start from next block
		if event.BlockNumber >= d.fromBlock.Uint64() {
			d.fromBlock = big.NewInt(int64(event.BlockNumber) + 1)
		}
	}

	// update block in db
	if d.fromBlock.Cmp(lastBlock) > 0 {
		database.SetBlockNumber(d.fromBlock.Int64())
	}

	return nil
}

// unpack a log
func (d *Dumper) unpack(log types.Log, ABI abi.ABI, out interface{}) error {
	// get event name from map with hash
	eventName := d.eventNameMap[log.Topics[0]]
	// get all topics
	indexed := d.indexedMap[log.Topics[0]]

	logger.Debug("log data: %x", log.Data)
	logger.Debug("log topics: ", log.Topics)

	// logger.Info("event name: ", eventName)
	logger.Debug("topics in map: ", indexed)

	// parse data
	logger.Debug("parse data, event name: ", eventName)
	err := ABI.UnpackIntoInterface(out, eventName, log.Data)
	if err != nil {
		return err
	}
	logger.Debug("unpack out(no topics):", out)

	// parse topic
	logger.Debug("parse topic")
	err = abi.ParseTopics(out, indexed, log.Topics[1:])
	if err != nil {
		return err
	}
	logger.Debug("unpack out(with topics):", out)

	return nil
}

// func recoverAddressFromTx(tx *types.Transaction) (common.Address, error) {
// 	return types.LatestSignerForChainID(tx.ChainId()).Sender(tx)
// }
