package dumper

import (
	"math/big"

	"github.com/gridprotocol/dumper/database"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type AddNodeEvent struct {
	Cp common.Address
	Id uint64

	Cpu struct {
		CpuPriceMon *big.Int
		CpuPriceSec *big.Int
		Model       string
	}

	Gpu struct {
		GpuPriceMon *big.Int
		GpuPriceSec *big.Int
		Model       string
	}

	Mem struct {
		MemPriceMon *big.Int
		MemPriceSec *big.Int
		Num         uint64
	}

	Disk struct {
		DiskPriceMon *big.Int
		DiskPriceSec *big.Int
		Num          uint64
	}
}

// unpack log data and store into db
func (d *Dumper) HandleAddNode(log types.Log) error {
	var out AddNodeEvent

	// abi0 = registry
	err := d.unpack(log, d.contractABI[0], &out)
	if err != nil {
		return err
	}

	// make node with data
	nodeInfo := database.Node{
		Address: out.Cp.Hex(),
		Id:      out.Id,

		CPUPrice: out.Cpu.CpuPriceSec,
		CPUModel: out.Cpu.Model,

		GPUPrice: out.Gpu.GpuPriceSec,
		GPUModel: out.Gpu.Model,

		MemPrice:    out.Mem.MemPriceSec,
		MemCapacity: int64(out.Mem.Num),

		DiskPrice:    out.Disk.DiskPriceSec,
		DiskCapacity: int64(out.Disk.Num),
	}

	logger.Info("store AddNode..")
	// store data
	err = nodeInfo.CreateNode()
	if err != nil {
		logger.Debug("store AddNode error: ", err.Error())
		return err
	}

	return nil
}
