package dumper

import (
	"math/big"
	"time"

	"github.com/gridprotocol/dumper/database"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type CreateOrderEvent struct {
	Cp  common.Address
	Id  uint64
	Nid uint64
	Act *big.Int
	Pro *big.Int
	Dur *big.Int
}

func (d *Dumper) HandleCreateOrder(log types.Log, from common.Address) error {
	var out CreateOrderEvent

	// abi1 = market
	err := d.unpack(log, d.contractABI[1], &out)
	if err != nil {
		return err
	}

	startTime := out.Act.Add(out.Act, out.Pro)
	endTime := startTime.Add(startTime, out.Dur)
	orderInfo := database.Order{
		User:         from.Hex(),
		Provider:     out.Cp.Hex(),
		Id:           out.Id,
		Nid:          out.Nid,
		ActivateTime: time.Unix(out.Act.Int64(), 0),
		StartTime:    time.Unix(startTime.Int64(), 0),
		EndTime:      time.Unix(endTime.Int64(), 0),
		Probation:    out.Pro.Int64(),
		Duration:     out.Dur.Int64(),
	}

	logger.Info("store create order..")
	err = orderInfo.CreateOrder()
	if err != nil {
		logger.Debug("store create order error: ", err.Error())
		return err
	}

	// get node info from db
	nodeInfo, err := database.GetNodeByCpAndId(orderInfo.Provider, orderInfo.Id)
	if err != nil {
		return err
	}

	// get profit info
	profitInfo, err := database.GetProfitByAddress(orderInfo.Provider)
	if err != nil {
		return err
	}

	// init profit info
	// (cpuPrice + gpuPrice + memPrice + diskPrice) * duration
	price := new(big.Int).Add(nodeInfo.CPUPrice, nodeInfo.GPUPrice)
	price.Add(price, nodeInfo.MemPrice)
	price.Add(price, nodeInfo.DiskPrice)
	price.Mul(price, big.NewInt(orderInfo.Duration))

	profitInfo.Profit.Add(profitInfo.Profit, price)
	if orderInfo.EndTime.Compare(profitInfo.EndTime) == 1 {
		profitInfo.EndTime = orderInfo.EndTime
	}

	return profitInfo.UpdateProfit()
}

type WithdrawEvent struct {
	Cp     common.Address
	Amount *big.Int
}

func (d *Dumper) HandleWithdraw(log types.Log) error {
	var out WithdrawEvent
	err := d.unpack(log, d.contractABI[1], &out)
	if err != nil {
		return err
	}

	profit, err := database.GetProfitByAddress(out.Cp.Hex())
	if err != nil {
		return err
	}

	profit.Balance.Sub(profit.Balance, out.Amount)
	profit.Nonce++
	return profit.UpdateProfit()
}
