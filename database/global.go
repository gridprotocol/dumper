package database

import (
	"time"

	"gorm.io/gorm"
)

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

// GetProviderCount 查询 Provider 表中的记录数量
func GetProviderCount() (int64, error) {
	var count int64

	// 查询 Provider 表中的记录数量
	if err := GlobalDataBase.Model(&Provider{}).Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

// get all nodes count
func GetNodeCount() (int64, error) {
	var count int64

	// 查询 node 表中的记录数量
	if err := GlobalDataBase.Model(&NodeStore{}).Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

// get all used nodes count
func GetNodeCountInOrders() (int64, error) {
	// 查询所有 Order 记录，并连接 NodeStore 表
	var count int64
	if err := GlobalDataBase.Model(&Order{}).
		Joins("JOIN node_stores ns ON ns.address = orders.provider AND ns.id = orders.nid").
		Group("ns.address, ns.id").
		Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

// get all nodes' mem count
func GetTotalMemCapacity() (int64, error) {
	var totalMemCapacity int64

	// 查询所有节点的 MemCapacity 总量
	if err := GlobalDataBase.Model(&NodeStore{}).Select("SUM(mem_capacity)").Scan(&totalMemCapacity).Error; err != nil {
		return 0, err
	}

	return totalMemCapacity, nil
}

// get all nodes' disk count
func GetTotalDiskCapacity() (int64, error) {
	var totalDiskCapacity int64

	// 查询所有节点的 MemCapacity 总量
	if err := GlobalDataBase.Model(&NodeStore{}).Select("SUM(disk_capacity)").Scan(&totalDiskCapacity).Error; err != nil {
		return 0, err
	}

	return totalDiskCapacity, nil
}

// GetTotalResources 查询所有节点的 MemCapacity 和 DiskCapacity 总量
func GetTotalResources() (int64, int64, error) {
	var result struct {
		TotalMemCapacity  int64 `gorm:"total_mem_capacity"`
		TotalDiskCapacity int64 `gorm:"total_disk_capacity"`
	}

	// 查询所有节点的 MemCapacity 和 DiskCapacity 总量
	if err := GlobalDataBase.Model(&NodeStore{}).
		Select("SUM(mem_capacity) AS total_mem_capacity, SUM(disk_capacity) AS total_disk_capacity").
		Scan(&result).Error; err != nil {
		return 0, 0, err
	}

	return result.TotalMemCapacity, result.TotalDiskCapacity, nil
}

// get all active orders' total mem
func GetTotalMemCapacityOfActivedOrders() (int64, error) {
	// 获取当前时间
	now := time.Now()

	// 查询所有 endtime 大于当前时间的 Order 记录，并连接 NodeStore 表
	var totalMemCapacity int64
	if err := GlobalDataBase.Model(&Order{}).
		Select("SUM(ns.mem_capacity)").
		Joins("JOIN node_stores ns ON ns.address = orders.provider AND ns.id = orders.nid").
		Where("orders.end > ?", now).
		Scan(&totalMemCapacity).Error; err != nil {
		return 0, err
	}

	return totalMemCapacity, nil
}

// get all active orders' total disk
func GetTotalDiskCapacityOfActivedOrders() (int64, error) {
	// 获取当前时间
	now := time.Now()

	// 查询所有 endtime 大于当前时间的 Order 记录，并连接 NodeStore 表
	var totalDiskCapacity int64
	if err := GlobalDataBase.Model(&Order{}).
		Select("SUM(ns.disk_capacity)").
		Joins("JOIN node_stores ns ON ns.address = orders.provider AND ns.id = orders.nid").
		Where("orders.end > ?", now).
		Scan(&totalDiskCapacity).Error; err != nil {
		return 0, err
	}

	return totalDiskCapacity, nil
}

func GetUsedResources() (int64, int64, error) {
	// 获取当前时间
	now := time.Now()

	// 查询所有 endtime 大于当前时间的 Order 记录，并连接 NodeStore 表
	var result struct {
		TotalMemCapacity  int64 `gorm:"total_mem_capacity"`
		TotalDiskCapacity int64 `gorm:"total_disk_capacity"`
	}

	if err := GlobalDataBase.Model(&Order{}).
		Select("SUM(ns.mem_capacity) AS total_mem_capacity, SUM(ns.disk_capacity) AS total_disk_capacity").
		Joins("JOIN node_stores ns ON ns.address = orders.provider AND ns.id = orders.nid").
		Where("orders.end > ?", now).
		Scan(&result).Error; err != nil {
		return 0, 0, err
	}

	return result.TotalMemCapacity, result.TotalDiskCapacity, nil
}
