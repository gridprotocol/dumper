package database

import (
	"math/big"

	"golang.org/x/xerrors"
)

type Node struct {
	Address string
	Id      uint64

	CPUPriceMon *big.Int
	CPUPriceSec *big.Int
	CPUModel    string
	CPUCore     uint64

	GPUPriceMon *big.Int
	GPUPriceSec *big.Int
	GPUModel    string

	MemPriceMon *big.Int
	MemPriceSec *big.Int
	MemCapacity int64

	DiskPriceMon *big.Int
	DiskPriceSec *big.Int
	DiskCapacity int64

	Exist bool
	Sold  bool
	Avail bool

	Online bool
}

type NodeStore struct {
	Address string `gorm:"primaryKey"`
	Id      uint64 `gorm:"primaryKey;autoIncrement:false"`

	CPUPriceMon string
	CPUPriceSec string
	CPUModel    string
	CPUCore     uint64

	GPUPriceMon string
	GPUPriceSec string
	GPUModel    string

	MemPriceMon string
	MemPriceSec string
	MemCapacity int64

	DiskPriceMon string
	DiskPriceSec string
	DiskCapacity int64

	Exist bool
	Sold  bool
	Avail bool

	Online bool
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
func ListAllNodesByUser(user string) ([]NodeAdaptor, error) {
	var nodes []NodeStore
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

	err = GlobalDataBase.Model(&NodeStore{}).Where("address IN (?) AND id IN (?)", providers, nids).Find(&nodes).Error
	if err != nil {
		return nil, err
	}

	nodeAdps := []NodeAdaptor{}
	// 适配node到nodeInProvider
	for _, n := range nodes {
		// compatible node to node_in
		node := NodeAdaptor{
			ID: n.Id,
			CP: n.Address,

			CPU: CPU{
				PriceMon: n.CPUPriceMon,
				PriceSec: n.CPUPriceSec,
				Model:    n.CPUModel,
				Core:     n.CPUCore,
			},
			GPU: GPU{
				PriceMon: n.GPUPriceMon,
				PriceSec: n.GPUPriceSec,
				Model:    n.GPUModel,
			},
			MEM: MEM{
				PriceMon: n.MemPriceMon,
				PriceSec: n.MemPriceSec,
				Num:      n.MemCapacity,
			},
			DISK: DISK{
				PriceMon: n.DiskPriceMon,
				PriceSec: n.DiskPriceSec,
				Num:      n.DiskCapacity,
			},

			Exist:  n.Exist,
			Sold:   n.Sold,
			Avail:  n.Avail,
			Online: n.Online,
		}

		// get order's appname with provider and nid
		var order Order
		result := GlobalDataBase.Where("provider = ? AND nid = ?", node.CP, node.ID).First(&order)
		if result.Error != nil {
			return nil, result.Error
		}
		node.AppName = order.AppName

		nodeAdps = append(nodeAdps, node)
	}

	return nodeAdps, nil
}

// set node exist
func SetExist(cp string, id uint64, set bool) error {
	// 更新 node_stores 表中相应节点的 exist 字段
	err := GlobalDataBase.Model(&NodeStore{}).Where("address = ? AND id = ?", cp, id).Update("exist", set).Error
	if err != nil {
		return err
	}

	return nil
}

// set node sold
func SetSold(cp string, id uint64, set bool) error {
	// 更新 node_stores 表中相应节点的 sold 字段
	err := GlobalDataBase.Model(&NodeStore{}).Where("address = ? AND id = ?", cp, id).Update("sold", set).Error
	if err != nil {
		return err
	}

	return nil
}

// set node avail
func SetAvail(cp string, id uint64, set bool) error {
	// 更新 node_stores 表中相应节点的 avail 字段
	err := GlobalDataBase.Model(&NodeStore{}).Where("address = ? AND id = ?", cp, id).Update("avail", set).Error
	if err != nil {
		return err
	}

	return nil
}

// set node online
func SetOnline(cp string, id uint64, set bool) error {
	// 更新 node_stores 表中相应节点的 online 字段
	err := GlobalDataBase.Model(&NodeStore{}).Where("address = ? AND id = ?", cp, id).Update("online", set).Error
	if err != nil {
		return err
	}

	return nil
}

func NodeToNodeStore(node Node) (NodeStore, error) {
	return NodeStore{
		Address: node.Address,
		Id:      node.Id,

		CPUPriceMon: node.CPUPriceMon.String(),
		CPUPriceSec: node.CPUPriceSec.String(),
		CPUModel:    node.CPUModel,
		CPUCore:     node.CPUCore,

		GPUPriceMon: node.GPUPriceMon.String(),
		GPUPriceSec: node.GPUPriceSec.String(),
		GPUModel:    node.GPUModel,

		MemPriceMon: node.MemPriceMon.String(),
		MemPriceSec: node.MemPriceSec.String(),
		MemCapacity: node.MemCapacity,

		DiskPriceMon: node.DiskPriceMon.String(),
		DiskPriceSec: node.DiskPriceSec.String(),
		DiskCapacity: node.DiskCapacity,

		Exist: node.Exist,
		Sold:  node.Sold,
		Avail: node.Avail,
	}, nil
}

func NodeStoreToNode(node NodeStore) (Node, error) {
	cpuPriceMon, ok := new(big.Int).SetString(node.CPUPriceMon, 10)
	if !ok {
		return Node{}, xerrors.Errorf("Failed to convert %s to BigInt", node.CPUPriceMon)
	}

	cpuPriceSec, ok := new(big.Int).SetString(node.CPUPriceSec, 10)
	if !ok {
		return Node{}, xerrors.Errorf("Failed to convert %s to BigInt", node.CPUPriceSec)
	}

	gpuPriceMon, ok := new(big.Int).SetString(node.GPUPriceMon, 10)
	if !ok {
		return Node{}, xerrors.Errorf("Failed to convert %s to BigInt", node.GPUPriceMon)
	}

	gpuPriceSec, ok := new(big.Int).SetString(node.GPUPriceSec, 10)
	if !ok {
		return Node{}, xerrors.Errorf("Failed to convert %s to BigInt", node.GPUPriceMon)
	}

	memPriceMon, ok := new(big.Int).SetString(node.MemPriceMon, 10)
	if !ok {
		return Node{}, xerrors.Errorf("Failed to convert %s to BigInt", node.MemPriceMon)
	}

	memPriceSec, ok := new(big.Int).SetString(node.MemPriceSec, 10)
	if !ok {
		return Node{}, xerrors.Errorf("Failed to convert %s to BigInt", node.MemPriceSec)
	}

	diskPriceMon, ok := new(big.Int).SetString(node.DiskPriceMon, 10)
	if !ok {
		return Node{}, xerrors.Errorf("Failed to convert %s to BigInt", node.DiskPriceMon)
	}

	diskPriceSec, ok := new(big.Int).SetString(node.DiskPriceSec, 10)
	if !ok {
		return Node{}, xerrors.Errorf("Failed to convert %s to BigInt", node.DiskPriceSec)
	}

	return Node{
		Address: node.Address,
		Id:      node.Id,

		CPUPriceMon: cpuPriceMon,
		CPUPriceSec: cpuPriceSec,
		CPUModel:    node.CPUModel,

		GPUPriceMon: gpuPriceMon,
		GPUPriceSec: gpuPriceSec,
		GPUModel:    node.GPUModel,

		MemPriceMon: memPriceMon,
		MemPriceSec: memPriceSec,
		MemCapacity: node.MemCapacity,

		DiskPriceMon: diskPriceMon,
		DiskPriceSec: diskPriceSec,
		DiskCapacity: node.DiskCapacity,

		Exist: node.Exist,
		Sold:  node.Sold,
		Avail: node.Avail,

		Online: node.Online,
	}, nil
}
