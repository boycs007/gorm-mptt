package tests

import mptt "github.com/boycs007/gorm-mptt"

type CustomTree struct {
	mptt.ModelBase
	Name     string        `gorm:"type:varchar(125);index:custom_tree_name" validate:"required"`
	Children []*CustomTree `gorm:"-"`
}

type Node struct {
	Name     string  `json:"name"`
	ParentID int     `json:"-"`
	Children []*Node `json:"children,omitempty"`
}
