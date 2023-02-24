package mptt

import "gorm.io/gorm"

// Rebuild 全部数据修正
func (db *Tree) Rebuild() error {
    var existsTreeIds []int
    err := db.Statement.DB.Model(db.Context.Node).Select("tree_id").Order("tree_id").
        Where("parent_id = 0").Scan(&existsTreeIds).Error
    if err != nil {
        return err
    }

    lastTreeId := 0
    for _, treeId := range existsTreeIds {
        if treeId == lastTreeId {
            continue
        }
        err = db.PartialRebuild(treeId)
        if err != nil {
            return err
        }
        lastTreeId = treeId
    }

    err = db.Statement.DB.Model(db.Context.Node).Select("tree_id").Order("tree_id").
        Where("parent_id = 0").Scan(&existsTreeIds).Error
    if err != nil {
        return err
    }

    expectTreeId := 1
    for _, treeId := range existsTreeIds {
        diff := treeId - expectTreeId
        if diff != 0 {
            err = db.Statement.DB.Model(db.Context.Node).Where("tree_id = ?", treeId).
                Update("tree_id", gorm.Expr("tree_id - ?", diff)).Error
            if err != nil {
                return err
            }
        }
        expectTreeId++
    }
    return nil
}

// PartialRebuild 当一棵树的秩序混乱了时，需要根据parent_id关系对树进行修正
func (db *Tree) PartialRebuild(treeID int) error {
    var rootPks []int // 有可能创建了多个相同treeId的根节点
    err := db.Statement.DB.Model(db.Context.Node).Select("id").
        Where("parent_id = 0 AND tree_id = ?", treeID).Scan(&rootPks).Error
    if err != nil {
        return err
    }
    if len(rootPks) == 0 {
        return nil
    }
    _, err = db.rebuildHelper(rootPks[0], 1, treeID, 1)
    if err != nil {
        return err
    }
    if len(rootPks) == 1 {
        return nil
    }
    for _, rootPk := range rootPks[1:] {
        newTreeId := db.getNextTreeId(db.Context.Node)
        _, err = db.rebuildHelper(rootPk, 1, newTreeId, 1)
    }
    return err
}

// 递归一个个修正，效率会很低，但是能确保正确性
func (db *Tree) rebuildHelper(pk, left, treeId, level int) (int, error) {
    right := left + 1
    var children []int
    // 以原有lft为序修正Tree
    err := db.Statement.DB.Model(db.Context.Node).
        Select("id").Where("parent_id = ?", pk).Order("lft").Scan(&children).Error
    if err != nil {
        return 0, err
    }
    for _, child := range children {
        right, err = db.rebuildHelper(child, right, treeId, level+1)
        if err != nil {
            return right + 1, err
        }
    }
    err = db.Statement.DB.Model(db.Context.Node).Where("id = ?", pk).
        Select("tree_id", "lft", "rght", "lvl", "parent_id").
        Updates(map[string]interface{}{"tree_id": treeId, "lft": left, "rght": right, "lvl": level}).Error
    return right + 1, err
}
