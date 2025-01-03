package database

import (
	"math/big"

	"golang.org/x/xerrors"
)

type Node struct {
	Address string
	Id      uint64

	CPUPrice *big.Int
	CPUModel string

	GPUPrice *big.Int
	GPUModel string

	MemPrice    *big.Int
	MemCapacity int64

	DiskPrice    *big.Int
	DiskCapacity int64
}

type NodeStore struct {
	Address string `gorm:"primaryKey"`
	Id      uint64 `gorm:"primaryKey;autoIncrement:false"`

	CPUPrice string
	CPUModel string

	GPUPrice string
	GPUModel string

	MemPrice    string
	MemCapacity int64

	DiskPrice    string
	DiskCapacity int64
}

func InitNode() error {
	return GlobalDataBase.AutoMigrate(&NodeStore{})
}

// store node info to db
func (n *Node) CreateNode() error {
	nodeStore, err := NodeToNodeStore(*n)
	if err != nil {
		return err
	}
	return GlobalDataBase.Create(&nodeStore).Error
}

// get node with cp and id
func GetNodeByCpAndId(cp string, id uint64) (Node, error) {
	var nodeStore NodeStore
	err := GlobalDataBase.Model(&NodeStore{}).Where("address = ? AND id = ?", cp, id).First(&nodeStore).Error
	if err != nil {
		return Node{}, err
	}

	return NodeStoreToNode(nodeStore)
}

// list all node by specify start and num of node
func ListAllNodes(start, num int) ([]NodeStore, error) {
	var nodeStores []NodeStore

	err := GlobalDataBase.Model(&NodeStore{}).Limit(num).Offset(start).Find(&nodeStores).Error
	if err != nil {
		return nil, err
	}

	return nodeStores, nil
}

// get node list of a cp
func ListAllNodesByCp(cp string) ([]NodeStore, error) {
	var nodeStores []NodeStore
	err := GlobalDataBase.Model(&NodeStore{}).Where("address = ?", cp).Find(&nodeStores).Error
	if err != nil {
		return nil, err
	}

	return nodeStores, nil
}

// ListAllNodesByUser 通过用户地址查询与之相关的所有节点列表
func ListAllNodesByUser(user string) ([]NodeStore, error) {
	var nodeStores []NodeStore
	var orders []Order

	// 首先查询 orders 表，获取所有与用户相关的订单
	err := GlobalDataBase.Where("user = ?", user).Find(&orders).Error
	if err != nil {
		return nil, err
	}

	// 然后根据 orders 表中的 provider 和 nid 查询 node_stores 表中的节点信息
	var providers []string
	var nids []int
	for _, order := range orders {
		providers = append(providers, order.Provider)
		nids = append(nids, int(order.Nid))
	}

	err = GlobalDataBase.Model(&NodeStore{}).Where("address IN (?) AND id IN (?)", providers, nids).Find(&nodeStores).Error
	if err != nil {
		return nil, err
	}

	return nodeStores, nil
}

func NodeToNodeStore(node Node) (NodeStore, error) {
	return NodeStore{
		Address: node.Address,
		Id:      node.Id,

		CPUPrice: node.CPUPrice.String(),
		CPUModel: node.CPUModel,

		GPUPrice: node.GPUPrice.String(),
		GPUModel: node.GPUModel,

		MemPrice:    node.MemPrice.String(),
		MemCapacity: node.MemCapacity,

		DiskPrice:    node.DiskPrice.String(),
		DiskCapacity: node.DiskCapacity,
	}, nil
}

func NodeStoreToNode(node NodeStore) (Node, error) {
	cpuPrice, ok := new(big.Int).SetString(node.CPUPrice, 10)
	if !ok {
		return Node{}, xerrors.Errorf("Failed to convert %s to BigInt", node.CPUPrice)
	}

	gpuPrice, ok := new(big.Int).SetString(node.GPUPrice, 10)
	if !ok {
		return Node{}, xerrors.Errorf("Failed to convert %s to BigInt", node.GPUPrice)
	}

	memPrice, ok := new(big.Int).SetString(node.MemPrice, 10)
	if !ok {
		return Node{}, xerrors.Errorf("Failed to convert %s to BigInt", node.MemPrice)
	}

	diskPrice, ok := new(big.Int).SetString(node.DiskPrice, 10)
	if !ok {
		return Node{}, xerrors.Errorf("Failed to convert %s to BigInt", node.DiskPrice)
	}

	return Node{
		Address: node.Address,
		Id:      node.Id,

		CPUPrice: cpuPrice,
		CPUModel: node.CPUModel,

		GPUPrice: gpuPrice,
		GPUModel: node.GPUModel,

		MemPrice:    memPrice,
		MemCapacity: node.MemCapacity,

		DiskPrice:    diskPrice,
		DiskCapacity: node.DiskCapacity,
	}, nil
}
