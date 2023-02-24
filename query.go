package mptt

import (
    "gorm.io/gorm"
    "math"
)

func (db *Tree) GetDescendantCount() int {
    base := db.Context.ModelBase
    return int(math.Floor(float64(base.Rght-base.Lft-1) / 2))
}

func (db *Tree) GetAncestors(outListPtr interface{}, ascending, includeSelf bool) error {
    order := "lft desc"
    if ascending {
        order = "lft asc"
    }
    base := db.Context.ModelBase
    tx := db.Statement.Where("tree_id = ?", base.TreeID)
    if includeSelf {
        tx = tx.Where("lft <= ?", base.Lft).Where("rght >= ?", base.Rght)
    } else {
        tx = tx.Where("lft < ?", base.Lft).Where("rght > ?", base.Rght)
    }
    return tx.Order(order).Find(outListPtr).Error
}

func (db *Tree) GetDescendants(outListPtr interface{}, includeSelf bool) error {
    base := db.Context.ModelBase
    tx := db.Statement.Where("tree_id = ?", base.TreeID)
    if includeSelf {
        tx = tx.Where("lft >= ?", base.Lft).Where("rght =< ?", base.Rght)
    } else {
        tx = tx.Where("lft > ?", base.Lft).Where("rght < ?", base.Rght)
    }
    return tx.Order("lft asc").Find(outListPtr).Error
}

func (db *Tree) GetFamily(outListPtr interface{}) error {
    base := db.Context.ModelBase
    return db.Statement.Where("tree_id = ?", base.TreeID).Where(
        db.Where("lft <= ?", base.Lft).Where("rght >= ?", base.Rght).Or(
            db.Where("lft > ?", base.Lft).Where("rght < ?", base.Rght),
        ),
    ).Order("lft asc").Find(outListPtr).Error
}

func (db *Tree) GetChildren(outListPtr interface{}) error {
    base := db.Context.ModelBase
    return db.Statement.
        Where("parent_id = ?", base.ID).
        Order("lft asc").
        Find(outListPtr).Error
}

func (db *Tree) GetLeafNodes(outListPtr interface{}) error {
    base := db.Context.ModelBase
    return db.Statement.
        Where("tree_id = ?", base.TreeID).
        Where("lft > ?", base.Lft).
        Where("rght < ?", base.Rght).
        Where("rght - lft = 1").
        Order("lft asc").
        Find(outListPtr).Error
}

func (db *Tree) GetSiblings(outListPtr interface{}, includeSelf bool) error {
    base := db.Context.ModelBase
    tx := db.Statement.
        Where("parent_id = ?", base.ParentID)
    if !includeSelf {
        tx = tx.Where("id != ?", base.ID)
    }
    return tx.Order("lft asc").Find(outListPtr).Error
}

func (db *Tree) nextSibling(tx *gorm.DB, base *MPTTModelBase) *gorm.DB {
    return tx.Where("parent_id = ?", base.ParentID).
        Where("rght < ?", base.Lft).
        Order("rght desc")
}

func (db *Tree) GetNextSibling(outPtr interface{}, conds ...interface{}) error {
    base := db.Context.ModelBase
    return db.Statement.DB.Scopes(func(tx *gorm.DB) *gorm.DB {
        return db.nextSibling(tx, base)
    }).First(outPtr, conds...).Error
}

func (db *Tree) previousSibling(tx *gorm.DB, base *MPTTModelBase) *gorm.DB {
    return tx.Where("parent_id = ?", base.ParentID).
        Where("rght < ?", base.Lft).
        Order("rght desc")
}

func (db *Tree) GetPreviousSibling(outPtr interface{}, conds ...interface{}) error {
    base := db.Context.ModelBase
    return db.Statement.DB.Scopes(func(tx *gorm.DB) *gorm.DB {
        return db.previousSibling(tx, base)
    }).First(outPtr, conds...).Error
}

func (db *Tree) GetRoot(outPtr interface{}) error {
    return db.RootNode(db.Context.ModelBase.TreeID, outPtr)
}

func (db *Tree) GetLevel() int {
    return db.Context.ModelBase.Lvl
}

func (db *Tree) IsChildNode() bool {
    return !db.IsRootNode()
}

func (db *Tree) IsRootNode() bool {
    return db.Context.ModelBase.ParentID == 0
}

func (db *Tree) IsLeafNode() bool {
    return db.GetDescendantCount() == 0
}

func (db *Tree) IsDescendantOf(other interface{}, includeSelf bool) bool {
    otherModel := db.getModelBase(other)
    currentModel := db.Context.ModelBase
    if includeSelf && otherModel.ID == currentModel.ID {
        return true
    }
    if otherModel.TreeID != currentModel.TreeID {
        return false
    }
    return currentModel.Lft > otherModel.Lft && currentModel.Rght < otherModel.Rght
}

func (db *Tree) IsAncestorOf(other interface{}, includeSelf bool) bool {
    otherModel := db.getModelBase(other)
    currentModel := db.Context.ModelBase
    if includeSelf && otherModel.ID == currentModel.ID {
        return true
    }
    if otherModel.TreeID != currentModel.TreeID {
        return false
    }
    return currentModel.Lft < otherModel.Lft && currentModel.Rght > otherModel.Rght
}

func (db *Tree) RootNodes(outListPtr interface{}) error {
    return db.Statement.Find(outListPtr, "parent_id = 0").Error
}

func (db *Tree) RootNode(treeID int, outPtr interface{}) error {
    return db.Statement.Where("parent_id = 0 AND tree_id = ?", treeID).Find(outPtr).Error
}
