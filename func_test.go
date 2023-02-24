package mptt

import (
    "reflect"
    "testing"
)

type TestModel struct {
    MPTTModelBase
    Name string `gorm:"index"`
}

var tree = Tree{}

func Test_getModelBase(t *testing.T) {
    m := TestModel{
        MPTTModelBase: MPTTModelBase{},
        Name:          "test",
    }
    tree.getModelBase(&m)
}

func Test_setModelBase(t *testing.T) {
    m := TestModel{
        MPTTModelBase: MPTTModelBase{},
        Name:          "test",
    }
    tree.setModelBase(&m, &MPTTModelBase{
        ParentID: 1,
        ID:       2,
        Lvl:      2,
        TreeID:   3,
        Lft:      2,
        Rght:     3,
    })
    if m.Lft != 2 {
        t.Error("error set.")
    }
}

func Test_reflectModel(t *testing.T) {
    m := TestModel{
        MPTTModelBase: MPTTModelBase{},
        Name:          "test",
    }
    node := reflect.New(reflect.TypeOf(m)).Interface()
    node1 := reflect.New(reflect.TypeOf(&m)).Interface()
    print(node)
    print(node1)
}
