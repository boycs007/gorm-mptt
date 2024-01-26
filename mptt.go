package mptt

import (
	"gorm.io/gorm"
)

type tree struct {
	*gorm.DB
	node      interface{}
	tableName string
	fields    *KeyFields
}

func (t *tree) GormDB() *gorm.DB {
	return t.DB
}

func (t *tree) getTableName() string {
	return t.Statement.Quote(t.tableName)
}

type treeOptions struct {
	specialTableName string
	keyColumns       KeyColumnFields
}

// ModelBase default mptt base model for user to embedded
type ModelBase struct {
	ID       int `gorm:"primaryKey" json:"id"`
	ParentID int `gorm:"default:0;index" json:"parent_id"`
	TreeID   int `gorm:"index" json:"tree_id"`
	Lvl      int `gorm:"index" json:"lvl"`
	Lft      int `gorm:"index" json:"lft"`
	Rght     int `gorm:"index" json:"rght"`
}

// Option ...
type Option func(options *treeOptions)

// KeyColumnFields for custom mptt model without embedded struct ModelBase
type KeyColumnFields struct {
	IDFieldName     string
	ParentFieldName string
	TreeIDFieldName string
	LeftFieldName   string
	RightFieldName  string
	LevelFieldName  string
}

// WithTableName custom table name
func WithTableName(name string) Option {
	return func(options *treeOptions) {
		options.specialTableName = name
	}
}

// WithAttrs custom model without Embedded Struct ModelBase
func WithAttrs(fields KeyColumnFields) Option {
	return func(options *treeOptions) {
		options.keyColumns = fields
	}
}

// NewTreeManager create mptt tree manager
func NewTreeManager(db *gorm.DB, modelPtr interface{}, opts ...Option) (TreeManager, error) {
	t := tree{
		DB:   db,
		node: modelPtr,
	}
	options := &treeOptions{
		keyColumns: KeyColumnFields{
			IDFieldName:     DefaultIDColumn,
			ParentFieldName: DefaultParentIDColumn,
			TreeIDFieldName: DefaultTreeIDColumn,
			LeftFieldName:   DefaultLeftColumn,
			RightFieldName:  DefaultRightColumn,
			LevelFieldName:  DefaultLevelColumn,
		},
	}

	for _, opt := range opts {
		opt(options)
	}
	err := t.Statement.ParseWithSpecialTableName(t.node, options.specialTableName)
	if err != nil {
		return nil, err
	}
	t.fields = &KeyFields{
		ID: KeyField{
			Attr:  ColumnIDAttr,
			Field: t.Statement.Schema.FieldsByName[options.keyColumns.IDFieldName],
		},
		Parent: KeyField{
			Attr:  ColumnParentIDAttr,
			Field: t.Statement.Schema.FieldsByName[options.keyColumns.ParentFieldName],
		},
		Tree: KeyField{
			Attr:  ColumnTreeIDAttr,
			Field: t.Statement.Schema.FieldsByName[options.keyColumns.TreeIDFieldName],
		},
		Left: KeyField{
			Attr:  ColumnLeftAttr,
			Field: t.Statement.Schema.FieldsByName[options.keyColumns.LeftFieldName],
		},
		Right: KeyField{
			Attr:  ColumnRightAttr,
			Field: t.Statement.Schema.FieldsByName[options.keyColumns.RightFieldName],
		},
		Level: KeyField{
			Attr:  ColumnLevelAttr,
			Field: t.Statement.Schema.FieldsByName[options.keyColumns.LevelFieldName],
		},
	}
	t.tableName = t.Statement.Table
	return &t, nil
}

type PositionEnum string

const (
	LastChild  PositionEnum = "last-child"
	FirstChild PositionEnum = "first-child"
	Left       PositionEnum = "left"
	Right      PositionEnum = "right"
)

func (t *tree) Node(node interface{}) TreeNode {
	newTree := &tree{
		DB:        t.DB,
		node:      node,
		tableName: t.tableName,
		fields:    t.fields,
	}
	return newTree
}
