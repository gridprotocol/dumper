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
	Core     uint64 `json:"core"`
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
type NodeAdaptor struct {
	ID   uint64 `json:"id"`
	CP   string `json:"cp"`
	CPU  CPU    `json:"cpu"`
	GPU  GPU    `json:"gpu"`
	MEM  MEM    `json:"mem"`
	DISK DISK   `json:"disk"`

	Exist  bool `json:"exist"`
	Sold   bool `json:"sold"`
	Avail  bool `json:"avail"`
	Online bool `json:"online"`

	AppName string `json:"appname"`
}

type ProviderAdaptor struct {
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

	Nodes []NodeAdaptor `json:"nodes"`
}

func InitProvider() error {
	return GlobalDataBase.AutoMigrate(&Provider{})
}

// store provider info to db
func (p *Provider) CreateProvider() error {
	return GlobalDataBase.Create(p).Error
}

// get cp info
func GetProviderByAddress(address string) (ProviderAdaptor, error) {
	var provider Provider

	// load provider
	err := GlobalDataBase.Model(&Provider{}).Where("address = ?", address).First(&provider).Error
	if err != nil {
		return ProviderAdaptor{}, err
	}

	// adapt provider
	var nodes []NodeStore
	// load nodes
	err = GlobalDataBase.Where("address = ?", provider.Address).Find(&nodes).Error
	if err != nil {
		return ProviderAdaptor{}, err
	}

	nodes_in := []NodeAdaptor{}
	// 适配node到nodeInProvider
	for _, n := range nodes {
		// compatible node to node_in
		node_in := NodeAdaptor{
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

		nodes_in = append(nodes_in, node_in)
	}

	// 将Provider和其Node列表添加到新的数据结构中
	providerAdp := ProviderAdaptor{
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
	}

	return providerAdp, nil
}

// list all providers with nodes
func ListAllProviders(start int, num int) ([]ProviderAdaptor, error) {
	var providers []Provider
	var providersWithNodes []ProviderAdaptor

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

		var nodes_in []NodeAdaptor
		// 适配node到nodeInProvider
		for _, n := range nodes {
			// compatible node to node_in
			node_in := NodeAdaptor{
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

			nodes_in = append(nodes_in, node_in)
		}

		// 将Provider和其Node列表添加到新的数据结构中
		providersWithNodes = append(providersWithNodes, ProviderAdaptor{
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
