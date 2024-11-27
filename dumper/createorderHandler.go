package dumper

import (
	"math/big"
	"time"

	"github.com/gridprotocol/platform-v2/database"

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

	return nil
}
