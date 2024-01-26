package tests

import (
	"fmt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"os"
)

var globalDb *gorm.DB

// GormInitWithSqlite 初始化Gorm测试库，并执行migrate.
func GormInitWithSqlite(tmpDBPath string) {
	if _, err := os.Stat(tmpDBPath); err == nil {
		if err := os.Remove(tmpDBPath); err != nil {
			fmt.Printf("fail to delete sqlite, path = %s", tmpDBPath)
			panic(err)
		}
	}
	db, err := gorm.Open(sqlite.Open(tmpDBPath),
		&gorm.Config{
			NamingStrategy: schema.NamingStrategy{
				SingularTable: true, // 设置单数表名
			},
			Logger: logger.Default.LogMode(logger.Info),
		})

	if err != nil {
		fmt.Printf("sqlite failed to connect database, got error %v", err)
		panic(err)
	}

	sqlDB, err := db.DB()
	if err == nil {
		err = sqlDB.Ping()
	}
	if err != nil {
		fmt.Printf("failed to connect database, got error %v", err)
	}
	globalDb = db
}

// RunMigrations 执行对应Models的Migrations的操作。
func RunMigrations(allModels ...interface{}) {
	if err := globalDb.AutoMigrate(allModels...); err != nil {
		fmt.Printf("Failed to auto migrate, but got error %v", err)
		os.Exit(1)
	}

	for _, m := range allModels {
		if !globalDb.Migrator().HasTable(m) {
			fmt.Printf("Failed to create table for %#v", m)
			os.Exit(1)
		}
	}
}

func init() {
	refreshDb()
}

func refreshDb() {
	GormInitWithSqlite("./test.db")
	RunMigrations(new(CustomTree))
}
