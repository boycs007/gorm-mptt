package mptt

import (
	"strings"

	"gorm.io/gorm"
)

// MPTT Table consts
const (
	TableName          = "[table_tree]"
	ColumnIDAttr       = "[id]"
	ColumnParentIDAttr = "[parent_id]"
	ColumnTreeIDAttr   = "[tree_id]"
	ColumnLeftAttr     = "[left]"
	ColumnRightAttr    = "[right]"
	ColumnLevelAttr    = "[level]"

	DefaultIDColumn       = "ID"
	DefaultParentIDColumn = "ParentID"
	DefaultTreeIDColumn   = "TreeID"
	DefaultLeftColumn     = "Lft"
	DefaultRightColumn    = "Rght"
	DefaultLevelColumn    = "Lvl"
)

// replacePlaceholder ... 占位符替换为具体的值
func (t *tree) replacePlaceholder(rawSql string) string {

	replacer := strings.NewReplacer(
		TableName, t.getTableName(),
		ColumnIDAttr, t.colID(),
		ColumnParentIDAttr, t.colParent(),
		ColumnTreeIDAttr, t.colTree(),
		ColumnLeftAttr, t.colLeft(),
		ColumnRightAttr, t.colRight(),
		ColumnLevelAttr, t.colLevel(),
	)
	return replacer.Replace(rawSql)
}

func (t *tree) colID(withoutQuote ...bool) string {
	if len(withoutQuote) > 0 && withoutQuote[0] {
		return t.fields.ID.DBName
	}
	return t.Statement.Quote(t.fields.ID.DBName)
}

func (t *tree) colParent(withoutQuote ...bool) string {
	if len(withoutQuote) > 0 && withoutQuote[0] {
		return t.fields.Parent.DBName
	}
	return t.Statement.Quote(t.fields.Parent.DBName)
}

func (t *tree) colTree(withoutQuote ...bool) string {
	if len(withoutQuote) > 0 && withoutQuote[0] {
		return t.fields.Tree.DBName
	}
	return t.Statement.Quote(t.fields.Tree.DBName)
}

func (t *tree) colLeft(withoutQuote ...bool) string {
	if len(withoutQuote) > 0 && withoutQuote[0] {
		return t.fields.Left.DBName
	}
	return t.Statement.Quote(t.fields.Left.DBName)
}

func (t *tree) colRight(withoutQuote ...bool) string {
	if len(withoutQuote) > 0 && withoutQuote[0] {
		return t.fields.Right.DBName
	}
	return t.Statement.Quote(t.fields.Right.DBName)
}

func (t *tree) colLevel(withoutQuote ...bool) string {
	if len(withoutQuote) > 0 && withoutQuote[0] {
		return t.fields.Level.DBName
	}
	return t.Statement.Quote(t.fields.Level.DBName)
}

func (t *tree) getNodeID(n interface{}) interface{} {
	return getFieldValue(n, t.fields.ID)
}

func (t *tree) getParentID(n interface{}) interface{} {
	return getFieldValue(n, t.fields.Parent)
}

func (t *tree) getLeft(n interface{}) int {
	return getIntFieldValue(n, t.fields.Left)
}

func (t *tree) getRight(n interface{}) int {
	return getIntFieldValue(n, t.fields.Right)
}

func (t *tree) getLevel(n interface{}) int {
	return getIntFieldValue(n, t.fields.Level)
}

func (t *tree) getTreeID(n interface{}) int {
	return getIntFieldValue(n, t.fields.Tree)
}

func (t *tree) setNodeID(n interface{}, value interface{}) {
	setFieldValue(n, t.fields.ID, value)
}

func (t *tree) setParentID(n interface{}, value interface{}) {
	setFieldValue(n, t.fields.Parent, value)
}

func (t *tree) setLeft(n interface{}, left int) {
	setFieldValue(n, t.fields.Left, left)
}

func (t *tree) setRight(n interface{}, right int) {
	setFieldValue(n, t.fields.Right, right)
}

func (t *tree) setLevel(n interface{}, level int) {
	setFieldValue(n, t.fields.Level, level)
}

func (t *tree) setTreeID(n interface{}, treeID int) {
	setFieldValue(n, t.fields.Tree, treeID)
}

func (t *tree) getNodeByID(id interface{}) (interface{}, error) {
	node := reflectNew(t.node)
	t.setNodeID(node, id)
	err := t.Model(node).First(node).Error
	return node, err
}

func (t *tree) getNextTreeId() int {
	var (
		treeId       int
		node         = reflectNew(t.node)
		treeIdDbName = t.colTree()
	)
	t.Statement.Select(treeIdDbName).
		Model(node).Order(treeIdDbName + " DESC").Limit(1).Scan(&treeId)
	return treeId + 1
}

func (t *tree) createTreeSpace(model interface{}, targetTreeId, num int) error {
	return t.Model(reflectNew(model)).
		Where(t.colTree()+" > ?", targetTreeId).
		Update(t.colTree(true), gorm.Expr(t.colTree()+" + ?", num)).Error
}
