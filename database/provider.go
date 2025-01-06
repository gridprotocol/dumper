package database

type Provider struct {
	Address string `gorm:"primarykey"`
	Name    string
	IP      string
	Domain  string
	Port    string
}

// for return json compatible
type CPU struct {
	PriceMon string `json:"priceMon"`
	PriceSec string `json:"priceSec"`
	Model    string `json:"model"`
	Core     int64  `json:"core"`
}
type GPU struct {
	PriceMon string `json:"priceMon"`
	PriceSec string `json:"priceSec"`
	Model    string `json:"model"`
}
type MEM struct {
	PriceMon string `json:"priceMon"`
	PriceSec string `json:"priceSec"`
	Num      int64  `json:"num"`
}
type DISK struct {
	PriceMon string `json:"priceMon"`
	PriceSec string `json:"priceSec"`
	Num      int64  `json:"num"`
}
type NodeInProvider struct {
	ID   uint64 `json:"id"`
	CP   string `json:"cp"`
	CPU  CPU    `json:"cpu"`
	GPU  GPU    `json:"gpu"`
	MEM  MEM    `json:"mem"`
	DISK DISK   `json:"disk"`

	Exist bool `json:"exist"`
	Sold  bool `json:"sold"`
	Avail bool `json:"avail"`
}

type ProviderWithNodes struct {
	Address string `gorm:"primarykey" json:"address"`
	Name    string `json:"name"`
	IP      string `json:"ip"`
	Domain  string `json:"domain"`
	Port    string `json:"port"`

	NNode uint64 `json:"nNode"`
	UNode uint64 `json:"uNode"`
	NMem  uint64 `json:"nMem"`
	UMem  uint64 `json:"uMem"`
	NDisk uint64 `json:"nDisk"`
	UDisk uint64 `json:"uDisk"`

	Nodes []NodeInProvider `json:"nodes"`
}

func InitProvider() error {
	return GlobalDataBase.AutoMigrate(&Provider{})
}

// store provider info to db
func (p *Provider) CreateProvider() error {
	return GlobalDataBase.Create(p).Error
}

func GetProviderByAddress(address string) (Provider, error) {
	var provider Provider
	err := GlobalDataBase.Model(&Provider{}).Where("address = ?", address).First(&provider).Error
	if err != nil {
		return Provider{}, err
	}

	return provider, nil
}

// list all providers with nodes
func ListAllProviders(start int, num int) ([]ProviderWithNodes, error) {
	var providers []Provider
	var providersWithNodes []ProviderWithNodes

	// 获取Provider列表
	err := GlobalDataBase.Model(&Provider{}).Limit(num).Offset(start).Find(&providers).Error
	if err != nil {
		return nil, err
	}

	// 为每个Provider获取对应的Node列表
	for _, provider := range providers {
		var nodes []NodeStore

		err = GlobalDataBase.Where("address = ?", provider.Address).Find(&nodes).Error
		if err != nil {
			return nil, err
		}

		var nodes_in []NodeInProvider
		// 适配node到nodeInProvider
		for _, n := range nodes {
			// compatible node to node_in
			node_in := NodeInProvider{
				ID: n.Id,
				CP: n.Address,

				CPU: CPU{
					PriceMon: n.CPUPrice,
					Model:    n.CPUModel,
					Core:     n.CPUCore,
				},
				GPU: GPU{
					PriceMon: n.GPUPrice,
					Model:    n.GPUModel,
				},
				MEM: MEM{
					PriceMon: n.MemPrice,
					Num:      n.MemCapacity,
				},
				DISK: DISK{
					PriceMon: n.DiskPrice,
					Num:      n.DiskCapacity,
				},

				Exist: n.Exist,
				Sold:  n.Sold,
				Avail: n.Avail,
			}

			nodes_in = append(nodes_in, node_in)
		}

		// 将Provider和其Node列表添加到新的数据结构中
		providersWithNodes = append(providersWithNodes, ProviderWithNodes{
			Address: provider.Address,
			Name:    provider.Name,
			IP:      provider.IP,
			Domain:  provider.Domain,
			Port:    provider.Port,

			NNode: 0,
			UNode: 0,
			NMem:  0,
			UMem:  0,
			NDisk: 0,
			UDisk: 0,

			Nodes: nodes_in,
		})
	}

	return providersWithNodes, nil
}
