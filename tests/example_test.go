package tests

import (
    "fmt"
    mptt "github.com/boycs007/gorm-mptt"
    . "github.com/smartystreets/goconvey/convey"
    "github.com/stretchr/testify/assert"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
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
    GormInitWithSqlite("./test.db")
    RunMigrations(new(CustomTree))
}

func Test_MutilTrees(t *testing.T) {
    manager := mptt.NewManager(globalDb)
    Convey("create root nodes", t, func() {

        roots := make([]*CustomTree, 0)
        for i := range make([]struct{}, 20) {
            node := &CustomTree{
                Name: fmt.Sprintf("RootNode%d", i),
            }
            err := manager.CreateNode(node)
            assert.Nil(t, err, "Create Node failed: %s", err)
            assert.Equal(t, node.TreeID, i+1)
            assert.Equal(t, node.Lft, 1)
            assert.Equal(t, node.Rght, 2)
            assert.Equal(t, node.Lvl, 1)
            roots = append(roots, node)
        }
        Convey("create last child", func() {

            depts := make([]*CustomTree, 0)

            for j := range make([]struct{}, 10) {
                subNode := &CustomTree{
                    Name: fmt.Sprintf("DeptNode%d", j),
                }
                err := manager.InsertNode(subNode, roots[4], mptt.LastChild, true)
                assert.Nil(t, err, "Insert Node failed: %s", err)
                assert.Equal(t, roots[4].Rght, 2+2*(j+1))
                assert.Equal(t, subNode.TreeID, roots[4].TreeID)
                assert.Equal(t, subNode.Lft, roots[4].Lft+(j*2)+1)

                depts = append(depts, subNode)
            }

            Convey("delete node", func() {
                err := manager.DeleteNode(roots[5])
                assert.Nil(t, err, "Delete Node failed: %s", err)
                err = manager.RefreshNode(roots[7])
                assert.Nil(t, err, "GetNode Node failed: %s", err)
                assert.Equal(t, 7, roots[7].TreeID)
                outPtr := &CustomTree{}
                err = manager.Node(roots[7]).GetRoot(outPtr)
                assert.Nil(t, err, "GetRoot Node failed: %s", err)
                assert.Equal(t, 8, outPtr.ID)
            })
        })
    })
}
