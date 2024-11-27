package database

type Provider struct {
	Address string `gorm:"primarykey"`
	Name    string
	IP      string
	Domain  string
	Port    string
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

func ListAllProviders(start int, num int) ([]Provider, error) {
	var providers []Provider

	err := GlobalDataBase.Model(&Provider{}).Limit(num).Offset(start).Find(&providers).Error
	if err != nil {
		return nil, err
	}

	return providers, nil
}
