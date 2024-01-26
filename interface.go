package mptt

import "gorm.io/gorm"

type TreeNode interface {
	GetAncestors(outListPtr interface{}, ascending, includeSelf bool) error
	GetDescendants(outListPtr interface{}, includeSelf bool) error
	GetFamily(outListPtr interface{}) error // my ancestors and my descendants
	GetChildren(outListPtr interface{}) error
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
}

// TreeManager ...
type TreeManager interface {
	GormDB() *gorm.DB
	CreateNode(node interface{}) error
	InsertNode(node, target interface{}, position PositionEnum) error
	MoveNode(node, target interface{}, position PositionEnum, refreshTarget ...bool) (bool, error)
	MoveNodeByID(nodeID, targetID interface{}, position PositionEnum) (bool, error)
	DeleteNode(n interface{}, doNotRefresh ...bool) error
	DeleteNodeByID(nodeID interface{}) error

	Rebuild() error
	PartialRebuild(treeID int) error

	RefreshNode(node interface{}) error
	Node(node interface{}) TreeNode
}
