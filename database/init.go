package database

import (
	// "database/sql"
	"os"
	"path/filepath"
	"time"

	"github.com/gridprotocol/dumper/logs"

	// "gorm.io/driver/mysql"

	// _ "github.com/go-sql-driver/mysql"
	"github.com/mitchellh/go-homedir"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var GlobalDataBase *gorm.DB
var logger = logs.Logger("database")

// init a gorm db with path
func InitDatabase(path string) error {
	// full path
	dir, err := homedir.Expand(path)
	if err != nil {
		return err
	}
	logger.Info("dumper db path: ", dir)

	// if dir not exist, make it
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		logger.Info("make dir")
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}
	}

	// dsn := "root@tcp(127.0.0.1:3306)/grid?charset=utf8mb4&parseTime=True&loc=Local"
	// mysqlDB, err := sql.Open("mysql", dsn)
	// if err != nil {
	// 	return err
	// }

	// open gorm db
	db, err := gorm.Open(sqlite.Open(filepath.Join(dir, "dumper.db")), &gorm.Config{})
	if err != nil {
		return err
	}

	// get sql db from gorm db
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	// 设置连接池中空闲连接的最大数量。
	sqlDB.SetMaxIdleConns(10)
	// 设置打开数据库连接的最大数量。
	sqlDB.SetMaxOpenConns(100)
	// 设置超时时间
	sqlDB.SetConnMaxLifetime(time.Second * 30)

	// ping db
	err = sqlDB.Ping()
	if err != nil {
		return err
	}

	// create all tables
	db.AutoMigrate(&Order{}, &ProfitStore{}, &BlockNumber{}, &Provider{}, &NodeStore{}, &GlobalStore{})
	GlobalDataBase = db

	// insert a record for global
	initialData := GlobalStore{
		Id:         0,
		CpNum:      0,
		NodeGlobal: 0,
		NodeUsed:   0,
		MemGlobal:  0,
		DiskGlobal: 0,
		MemUsed:    0,
		DiskUsed:   0,
	}
	// 尝试查找记录
	var existingData GlobalStore
	err = GlobalDataBase.First(&existingData, initialData.Id).Error
	if err == gorm.ErrRecordNotFound {
		// 记录不存在，插入新记录
		err = GlobalDataBase.Create(&initialData).Error
		if err != nil {
			panic("failed to insert initial data")
		}
	} else if err != nil {
		// 其他错误
		panic("failed to check for existing data")
	}

	logger.Info("init database success")

	return nil
}

func RemoveDataBase(path string) error {
	dir, err := homedir.Expand(path)
	if err != nil {
		return err
	}

	databasePath := filepath.Join(dir, "dumper.db")
	if _, err := os.Stat(databasePath); os.IsExist(err) {
		if err := os.Remove(databasePath); err != nil {
			return err
		}
	}

	return nil
}
