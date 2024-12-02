package database

import (
	"math/big"
	"time"

	"golang.org/x/xerrors"
)

type Profit struct {
	Address string `gorm:"primarykey"` // CPU/GPU供应商ID
	// ID       int64
	Balance  *big.Int  // 余额
	Profit   *big.Int  // 分润值
	Penalty  *big.Int  // 惩罚值
	LastTime time.Time // 上次更新时间
	EndTime  time.Time // 可以取出全部分润值时间
	Nonce    uint64
}

type ProfitStore struct {
	Address  string    `gorm:"primarykey"` // CPU/GPU供应商ID
	Balance  string    // 余额
	Profit   string    // 分润值
	Penalty  string    // 惩罚值
	LastTime time.Time // 上次更新时间
	EndTime  time.Time // 可以取出全部分润值时间
}

func InitProfit() error {
	return GlobalDataBase.AutoMigrate(&Profit{})
}

func (p *Profit) CreateProfit() error {
	ps := &ProfitStore{
		Address:  p.Address,
		Balance:  p.Balance.String(),
		Profit:   p.Profit.String(),
		Penalty:  p.Penalty.String(),
		LastTime: p.LastTime,
		EndTime:  p.EndTime,
	}
	return GlobalDataBase.Create(ps).Error
}

func (p *Profit) UpdateProfit() error {
	ps := &ProfitStore{
		Address:  p.Address,
		Balance:  p.Balance.String(),
		Profit:   p.Profit.String(),
		Penalty:  p.Penalty.String(),
		LastTime: p.LastTime,
		EndTime:  p.EndTime,
	}

	return GlobalDataBase.Model(&ProfitStore{}).Where("address = ?", p.Address).Save(ps).Error
}

func GetProfitByAddress(address string) (Profit, error) {
	var ps ProfitStore
	err := GlobalDataBase.Model(&ProfitStore{}).Where("address = ?", address).First(&ps).Error
	if err != nil {
		return Profit{}, err
	}

	var profit = Profit{
		Address:  ps.Address,
		LastTime: ps.LastTime,
		EndTime:  ps.EndTime,
	}

	_, ok := profit.Balance.SetString(ps.Balance, 10)
	if !ok {
		return Profit{}, xerrors.Errorf("Balance %s is not in decimal format", ps.Balance)
	}

	_, ok = profit.Profit.SetString(ps.Profit, 10)
	if !ok {
		return Profit{}, xerrors.Errorf("Profit %s is not in decimal format", ps.Profit)
	}

	_, ok = profit.Penalty.SetString(ps.Penalty, 10)
	if !ok {
		return Profit{}, xerrors.Errorf("Penalty %s is not in decimal format", ps.Penalty)
	}

	return profit, nil
}

var blockNumberKey = "block_number_key"

type BlockNumber struct {
	BlockNumberKey string `gorm:"primarykey;column:key"`
	BlockNumber    int64
}

func SetBlockNumber(blockNumber int64) error {
	var daBlockNumber = BlockNumber{
		BlockNumberKey: blockNumberKey,
		BlockNumber:    blockNumber,
	}
	return GlobalDataBase.Save(&daBlockNumber).Error
}

func GetBlockNumber() (int64, error) {
	var blockNumber BlockNumber
	err := GlobalDataBase.Model(&BlockNumber{}).First(&blockNumber).Error

	return blockNumber.BlockNumber, err
}
