package database

import (
	"fmt"
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
	Status       uint8
	AppName      string
}

func InitOrder() error {
	return GlobalDataBase.AutoMigrate(&Order{})
}

// store order info to db
func (o *Order) CreateOrder() error {
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

type OrderAdaptor struct {
	ID         uint64 `json:"id"`
	User       string `json:"user"`
	Provider   string `json:"provider"`
	Nid        uint64 `json:"node_id"`
	AppName    string `json:"appName"`
	Remain     string `json:"remain"`
	Remu       string `json:"remuneration"`
	ActiveTime int64  `json:"activeTime"`
	LastSettle int64  `json:"lastSettleTime"`
	Probation  int64  `json:"probation"`
	Duration   int64  `json:"duration"`
	// 0-not exist 1-unactive 2-active 3-cancelled 4-completed
	Status uint8 `json:"status"`
}

// user's orders
func ListAllOrderByUser(address string) ([]OrderAdaptor, error) {
	var orders []Order
	var ordersAdaptor []OrderAdaptor

	err := GlobalDataBase.Model(&Order{}).Where("user = ?", address).Find(&orders).Error
	if err != nil {
		return nil, err
	}

	for _, o := range orders {
		adp := OrderAdaptor{
			ID:         o.Id,
			User:       o.User,
			Provider:   o.Provider,
			Nid:        o.Nid,
			AppName:    "",
			Remain:     "",
			Remu:       "",
			ActiveTime: o.ActivateTime.Unix(),
			LastSettle: 0,
			Probation:  o.Probation,
			Duration:   o.Duration,
			Status:     o.Status,
		}
		ordersAdaptor = append(ordersAdaptor, adp)
	}

	return ordersAdaptor, nil
}

// provider's orders
func ListAllOrderByProvider(address string) ([]OrderAdaptor, error) {
	var orders []Order
	var ordersAdaptor []OrderAdaptor

	err := GlobalDataBase.Model(&Order{}).Where("provider = ?", address).Find(&orders).Error
	if err != nil {
		return nil, err
	}

	for _, o := range orders {
		adp := OrderAdaptor{
			ID:         o.Id,
			User:       o.User,
			Provider:   o.Provider,
			Nid:        o.Nid,
			AppName:    "",
			Remain:     "",
			Remu:       "",
			ActiveTime: o.ActivateTime.Unix(),
			LastSettle: 0,
			Probation:  o.Probation,
			Duration:   o.Duration,
			Status:     o.Status,
		}
		ordersAdaptor = append(ordersAdaptor, adp)
	}

	return ordersAdaptor, nil
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
	memFeeSec := new(big.Int).Mul(new(big.Int).SetInt64(node.MemCapacity), node.MemPriceSec)
	diskFeeSec := new(big.Int).Mul(new(big.Int).SetInt64(node.DiskCapacity), node.DiskPriceSec)

	totalPrice := new(big.Int).Add(node.CPUPriceSec, node.GPUPriceSec)
	totalPrice.Add(totalPrice, memFeeSec)
	totalPrice.Add(totalPrice, diskFeeSec)
	totalPrice.Mul(totalPrice, new(big.Int).SetInt64(order.Duration))

	// return order fee
	return totalPrice, nil
}

// set order status
func SetOrderStatus(oid uint64, st uint64) error {
	// 更新 node_stores 表中相应节点的 sold 字段
	err := GlobalDataBase.Model(&Order{}).Where("id = ?", oid).Update("status", st).Error
	if err != nil {
		return err
	}

	return nil
}

// check provider orders, if order is end, set status=4, and set node sold=false
func CheckProviderOrders(provider string) error {

	// 执行原生 SQL 更新订单状态
	result := GlobalDataBase.Exec("UPDATE orders SET status=4 WHERE provider= ? AND end < ?", provider, time.Now())

	// 检查并返回错误
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func UpdateOrderAndNodeStatus(provider string) error {
	// 获取当前时间
	now := time.Now()

	// 查询所有 endtime 小于当前时间且 provider 匹配的 Order 记录
	var orders []Order
	if err := GlobalDataBase.Model(&Order{}).
		Where("provider = ? AND end < ?", provider, now).
		Find(&orders).Error; err != nil {
		return fmt.Errorf("error fetching orders: %w", err)
	}

	// 更新这些 Order 记录的 status 为 4
	if err := GlobalDataBase.Model(&Order{}).
		Where("provider = ? AND end < ?", provider, now).
		Update("status", 4).Error; err != nil {
		return fmt.Errorf("error updating orders: %w", err)
	}

	// 更新这些 Order 记录对应的 NodeStore 记录的 sold 状态为 false
	for _, order := range orders {
		if err := GlobalDataBase.Model(&NodeStore{}).
			Where("address = ? AND id = ?", order.Provider, order.Nid).
			Update("sold", false).Error; err != nil {
			return fmt.Errorf("error updating node: %w", err)
		}
	}

	return nil
}
