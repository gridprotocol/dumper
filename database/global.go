package database

import "gorm.io/gorm"

type GlobalStore struct {
	Id uint64 `gorm:"primaryKey;autoIncrement:false"`

	CpNum      int64 `json:"cpNumber"`
	NodeGlobal int64 `json:"nodeGlobal"`
	NodeUsed   int64 `json:"nodeUsed"`
	MemGlobal  int64 `json:"memGlobal"`
	DiskGlobal int64 `json:"diskGlobal"`
	MemUsed    int64 `json:"memUsed"`
	DiskUsed   int64 `json:"diskUsed"`
}

// create global table
func InitGlobal() error {
	return GlobalDataBase.AutoMigrate(&GlobalStore{})
}

// store global info to db
func (g *GlobalStore) CreateGlobal() error {
	return GlobalDataBase.Create(&g).Error
}

// get global info
func GetGlobal() (GlobalStore, error) {
	var global GlobalStore
	err := GlobalDataBase.Model(&GlobalStore{}).Where("id = ?", 0).First(&global).Error
	if err != nil {
		return GlobalStore{}, err
	}

	return global, nil
}

// IncCp 累加 GlobalStore 表中的 CpNum 字段
func IncCp() error {
	// 假设 Id 为 0 的记录是需要更新的记录
	result := GlobalDataBase.Model(&GlobalStore{}).Where("id = ?", 0).UpdateColumn("cp_num", gorm.Expr("cp_num + ?", 1))
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// accu node resource
func IncNode(mem, disk int64) error {
	// 假设 Id 为 0 的记录是需要更新的记录
	result := GlobalDataBase.Model(&GlobalStore{}).Where("id = ?", 0).UpdateColumns(map[string]interface{}{
		"node_global": gorm.Expr("node_global + ?", 1),
		"mem_global":  gorm.Expr("mem_global + ?", mem),
		"disk_global": gorm.Expr("disk_global + ?", disk),
	})
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// increase used resource when createorder
func IncUsed(mem, disk int64) error {
	// 假设 Id 为 0 的记录是需要更新的记录
	result := GlobalDataBase.Model(&GlobalStore{}).Where("id = ?", 0).UpdateColumns(map[string]interface{}{
		"node_used": gorm.Expr("node_used + ?", 1),
		"mem_used":  gorm.Expr("mem_used + ?", mem),
		"disk_used": gorm.Expr("disk_used + ?", disk),
	})
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// decrease used resource when order en
func DecUsed(mem, disk int64) error {
	// 假设 Id 为 0 的记录是需要更新的记录
	result := GlobalDataBase.Model(&GlobalStore{}).Where("id = ?", 0).UpdateColumns(map[string]interface{}{
		"node_used": gorm.Expr("node_used - ?", 1),
		"mem_used":  gorm.Expr("mem_used - ?", mem),
		"disk_used": gorm.Expr("disk_used - ?", disk),
	})
	if result.Error != nil {
		return result.Error
	}
	return nil
}
