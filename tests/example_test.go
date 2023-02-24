package tests

import (
    "fmt"
    mptt "github.com/boycs007/gorm-mptt"
    "github.com/stretchr/testify/assert"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
    "gorm.io/gorm/schema"
    "os"
    "testing"
)

var globalDb *gorm.DB

type CustomTree struct {
    mptt.MPTTModelBase
    Name string `gorm:"type:varchar(125);index:custom_tree_name" validate:"required"`
}

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
    GormInitWithSqlite("./test.db")
    RunMigrations(new(CustomTree))
}

func Test_CreateNodes(t *testing.T) {
    manager := mptt.NewManager(globalDb)
    err := manager.CreateNode(&CustomTree{
        Name: "RootNode",
    })
    assert.Nil(t, err, "Create Node failed: %s", err)
}
