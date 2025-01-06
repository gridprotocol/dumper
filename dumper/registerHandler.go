package dumper

import (
	"math/big"
	"time"

	"github.com/gridprotocol/dumper/database"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type RegisterEvent struct {
	Cp     common.Address
	Name   string
	Ip     string
	Domain string
	Port   string
}

// parse a register log
func (d *Dumper) HandleRegister(log types.Log) error {
	var out RegisterEvent
	// abi0 - registry
	err := d.unpack(log, d.contractABI[0], &out)
	if err != nil {
		return err
	}

	// get log data
	providerInfo := database.Provider{
		Address: out.Cp.Hex(),
		Name:    out.Name,
		IP:      out.Ip,
		Domain:  out.Domain,
		Port:    out.Port,
	}

	// save data into db
	logger.Info("store register..")
	err = providerInfo.CreateProvider()
	if err != nil {
		logger.Debug("store register error: ", err.Error())
		return err
	}

	now := time.Now()
	profitInfo := database.Profit{
		Address:  out.Cp.Hex(),
		Balance:  big.NewInt(0),
		Profit:   big.NewInt(0),
		Penalty:  big.NewInt(0),
		LastTime: now,
		EndTime:  now,
	}
	return profitInfo.CreateProfit()
}
