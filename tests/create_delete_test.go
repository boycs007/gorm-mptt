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
        for i := range make([]struct{}, 10) {
            node := &CustomTree{
                Name: fmt.Sprintf("RootNode%d", i+1),
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

            for j := range make([]struct{}, 5) {
                subNode := &CustomTree{
                    Name: fmt.Sprintf("DeptNode%d", j+1),
                }
                err := manager.InsertNode(subNode, roots[4], mptt.LastChild, true)
                assert.Nil(t, err, "Insert Node failed: %s", err)
                assert.Equal(t, roots[4].Rght, 2+2*(j+1))
                assert.Equal(t, subNode.TreeID, roots[4].TreeID)
                assert.Equal(t, subNode.Lft, roots[4].Lft+(j*2)+1)

                depts = append(depts, subNode)
            }

            Convey("create first child", func() {
                secondNode := &CustomTree{
                    Name: "SecondGroup",
                }
                err := manager.InsertNode(secondNode, depts[3], mptt.FirstChild, true)
                assert.Nil(t, err, "Insert Node failed: %s", err)

                firstNode := &CustomTree{
                    Name: "FirstGroup",
                }
                err = manager.InsertNode(firstNode, depts[3], mptt.FirstChild, true)
                assert.Nil(t, err, "Insert Node failed: %s", err)

                err = manager.RefreshNode(secondNode)
                assert.Nil(t, err)
                assert.Equal(t, depts[3].Lft+3, secondNode.Lft)
                assert.Equal(t, depts[3].Lft+4, secondNode.Rght)
                assert.Equal(t, depts[3].Lvl+1, secondNode.Lvl)

                firstDept := &CustomTree{
                    Name: "FirstDept",
                }
                err = manager.InsertNode(firstDept, roots[4], mptt.FirstChild, true)
                assert.Nil(t, err, "Insert Node failed: %s", err)

                _ = manager.RefreshNode(roots[2])

                err = manager.InsertNode(&CustomTree{
                    Name: "InsertLeftRoot",
                }, roots[2], mptt.Left, true)
                assert.Nil(t, err, "Insert Node failed: %s", err)

                _ = manager.RefreshNode(depts[2])

                err = manager.InsertNode(&CustomTree{
                    Name: "InsertLeftDept",
                }, depts[2], mptt.Left, true)
                assert.Nil(t, err, "Insert Node failed: %s", err)

                Convey("delete node", func() {
                    err := manager.DeleteNode(roots[5])
                    assert.Nil(t, err, "Delete Node failed: %s", err)
                    err = manager.RefreshNode(roots[7])
                    assert.Nil(t, err, "GetNode Node failed: %s", err)
                    assert.Equal(t, 8, roots[7].TreeID)
                    outPtr := &CustomTree{}
                    err = manager.Node(roots[7]).GetRoot(outPtr)
                    assert.Nil(t, err, "GetRoot Node failed: %s", err)
                    assert.Equal(t, 8, outPtr.ID)

                    err = manager.DeleteNode(depts[1])
                    assert.Nil(t, err, "GetRoot Node failed: %s", err)
                })
            })

        })
    })
}
