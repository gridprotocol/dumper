package database

import (
	"math/big"
	"time"
)

type Order struct {
	Id           uint64 // order id
	User         string
	Provider     string
	Nid          uint64    // node id
	ActivateTime time.Time `gorm:"column:activate"`
	StartTime    time.Time `gorm:"column:start"`
	EndTime      time.Time `gorm:"column:end"`
	Probation    int64
	Duration     int64
}

func InitOrder() error {
	return GlobalDataBase.AutoMigrate(&Order{})
}

// store order info to db
func (o *Order) CreateOrder() error {
	o.StartTime = o.ActivateTime.Add(time.Duration(o.Probation) * time.Second)
	o.EndTime = o.StartTime.Add(time.Duration(o.Duration) * time.Second)
	return GlobalDataBase.Create(o).Error
}

// get order by order id
func GetOrderById(id uint64) (Order, error) {
	var order Order
	err := GlobalDataBase.Model(&Order{}).Where("id = ?", id).Last(&order).Error
	if err != nil {
		return Order{}, err
	}

	return order, nil
}

// get order list of an user
func GetOrdersByUser(user string) ([]Order, error) {
	var orders []Order

	err := GlobalDataBase.Model(&Order{}).Where("user = ?", user).Find(&orders).Error
	if err != nil {
		return nil, err
	}

	return orders, nil
}

// get orders count by provider address
func GetOrderCount(address string) (int64, error) {
	var cnt int64
	err := GlobalDataBase.Model(&Order{}).Where("provider = ?", address).Count(&cnt).Error
	if err != nil {
		return -1, err
	}

	return cnt, nil
}

func ListAllActivedOrder() ([]Order, error) {
	var now = time.Now()
	var orders []Order
	err := GlobalDataBase.Model(&Order{}).Where("start < ? AND end > ?", now, now).Find(&orders).Error
	if err != nil {
		return nil, err
	}

	return orders, nil
}

// user's orders
func ListAllOrderByUser(address string) ([]Order, error) {
	var orders []Order
	err := GlobalDataBase.Model(&Order{}).Where("user = ?", address).Find(&orders).Error
	if err != nil {
		return nil, err
	}

	return orders, nil
}

// user's active orders
func ListAllActivedOrderByUser(address string) ([]Order, error) {
	var now = time.Now()
	var orders []Order
	err := GlobalDataBase.Model(&Order{}).Where("user = ? AND start < ? AND end > ?", address, now, now).Find(&orders).Error
	//err := GlobalDataBase.Model(&Order{}).Where("user = ?", address).Find(&orders).Error
	if err != nil {
		return nil, err
	}

	return orders, nil
}

func ListAllOrderedProvider(user string) ([]Provider, error) {
	var now = time.Now()
	var provider []Provider

	err := GlobalDataBase.Model(&Order{}).Where("user = ? AND start < ? AND end > ?", user, now, now).
		//err := GlobalDataBase.Model(&Order{}).Where("user = ?", user).
		Joins("left join providers on orders.provider = providers.address").
		Select("address, name, ip,domain,port").Find(&provider).Error
	if err != nil {
		return nil, err
	}

	return provider, nil
}

// calc the fee of an order by id
func CalcOrderFee(id uint64) (*big.Int, error) {
	// get order info by id
	order, err := GetOrderById(id)
	if err != nil {
		return nil, err
	}

	// get node info by cp and nid
	node, err := GetNodeByCpAndId(order.Provider, order.Nid)
	if err != nil {
		return nil, err
	}

	logger.Debug("node: ", node)

	// calc order fee
	memFeeSec := new(big.Int).Mul(new(big.Int).SetInt64(node.MemCapacity), node.MemPrice)
	diskFeeSec := new(big.Int).Mul(new(big.Int).SetInt64(node.DiskCapacity), node.DiskPrice)

	totalPrice := new(big.Int).Add(node.CPUPrice, node.GPUPrice)
	totalPrice.Add(totalPrice, memFeeSec)
	totalPrice.Add(totalPrice, diskFeeSec)
	totalPrice.Mul(totalPrice, new(big.Int).SetInt64(order.Duration))

	// return order fee
	return totalPrice, nil
}
