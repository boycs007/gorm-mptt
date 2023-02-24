package mptt

import "gorm.io/gorm"

type PositionEnum string

const (
    LastChild  PositionEnum = "last-child"
    FirstChild PositionEnum = "first-child"
    Left       PositionEnum = "left"
    Right      PositionEnum = "right"
)

// MPTTModelBase ...
type MPTTModelBase struct {
    ID       int `gorm:"primaryKey" json:"id"`
    ParentID int `gorm:"default:0;index" json:"parent_id"`
    TreeID   int `gorm:"index" json:"tree_id"`
    Lvl      int `gorm:"index" json:"lvl"`
    Lft      int `gorm:"index" json:"lft"`
    Rght     int `gorm:"index" json:"rght"`
}

type TreeContext struct {
    Node      interface{}
    ModelBase *MPTTModelBase
}

type Tree struct {
    *gorm.DB
    Context TreeContext
}

func (db *Tree) Node(node interface{}) TreeManager {
    db.Context = db.getContext(node)
    return db
}

func (db *Tree) innerNode(n interface{}) *Tree {
    t := &Tree{
        DB:      db.Statement.DB,
        Context: db.getContext(n),
    }
    return t
}

func NewManager(db *gorm.DB) TreeManager {
    t := Tree{
        DB: db,
    }
    return &t
}

// TreeManager ...
type TreeManager interface {
    // CreateNode TODO delete
    CreateNode(node interface{}) error
    InsertNode(node, target interface{}, position PositionEnum, refreshTarget bool) error
    MoveNode(node, target interface{}, position PositionEnum) (bool, error)
    DeleteNode(n interface{}) error

    Rebuild() error
    PartialRebuild(treeID int) error

    Node(node interface{}) TreeManager

    // the interface below should call Node first

    GetAncestors(outListPtr interface{}, ascending, includeSelf bool) error
    GetFamily(outListPtr interface{}) error
    GetChildren(outListPtr interface{}) error
    GetDescendants(outListPtr interface{}, includeSelf bool) error
    GetLeafNodes(outListPtr interface{}) error
    GetSiblings(outListPtr interface{}, includeSelf bool) error
    GetNextSibling(outPtr interface{}, conds ...interface{}) error
    GetPreviousSibling(outPtr interface{}, conds ...interface{}) error
    GetRoot(outPtr interface{}) error
    GetDescendantCount() int
    GetLevel() int
    IsChildNode() bool
    IsLeafNode() bool
    IsRootNode() bool
    IsDescendantOf(other interface{}, includeSelf bool) bool
    IsAncestorOf(other interface{}, includeSelf bool) bool

    RootNodes(outListPtr interface{}) error
    RootNode(treeID int, outPtr interface{}) error
}
