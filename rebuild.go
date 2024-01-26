package mptt

import (
	"gorm.io/gorm"
)

// Rebuild 全部数据修正
func (t *tree) Rebuild() error {
	var existsTreeIds []int
	emptyNode := reflectNew(t.node)
	err := t.Model(emptyNode).Select(t.colTree()).Order(t.colTree()+" ASC").
		Where(t.colParent()+" = ?", t.getParentID(emptyNode)).Scan(&existsTreeIds).Error
	if err != nil {
		return err
	}

	lastTreeId := 0
	for _, treeId := range existsTreeIds {
		if lastTreeId != 0 && treeId == lastTreeId {
			continue
		}
		err = t.PartialRebuild(treeId)
		if err != nil {
			return err
		}
		existsTreeIds = append(existsTreeIds, treeId)
		lastTreeId = treeId
	}

	err = t.Model(emptyNode).Select(t.colTree()).Order(t.colTree()+" ASC").
		Where(t.colParent()+" = ?", t.getParentID(emptyNode)).Scan(&existsTreeIds).Error
	if err != nil {
		return err
	}

	expectTreeId := 1
	for _, treeId := range existsTreeIds {
		diff := treeId - expectTreeId
		if diff != 0 {
			err = t.Model(emptyNode).Select(t.colTree()).Where(t.colTree()+" = ?", treeId).
				Update(t.colTree(), gorm.Expr(t.colTree()+" - ?", diff)).Error
			if err != nil {
				return err
			}
		}
		expectTreeId++
	}
	return nil
}

// PartialRebuild 当一棵树的秩序混乱了时，需要根据parent_id关系对树进行修正
func (t *tree) PartialRebuild(treeID int) error {
	var rootPks []interface{} // 有可能创建了多个相同treeId的根节点
	emptyNode := reflectNew(t.node)
	whereSql := t.replacePlaceholder("[parent_id] = ? AND [tree_id] = ?")
	err := t.Model(emptyNode).Select(t.colID()).
		Where(whereSql, t.getParentID(emptyNode), treeID).Scan(&rootPks).Error
	if err != nil {
		return err
	}
	if len(rootPks) == 0 {
		return nil
	}
	if treeID == 0 {
		treeID = t.getNextTreeId()
	}
	_, err = t.rebuildHelper(rootPks[0], 1, treeID, 1)
	if err != nil {
		return err
	}
	if len(rootPks) == 1 {
		return nil
	}
	for _, rootPk := range rootPks[1:] {
		newTreeId := t.getNextTreeId()
		_, err = t.rebuildHelper(rootPk, 1, newTreeId, 1)
	}
	return err
}

// 递归一个个修正，效率会很低，但是能确保正确性
func (t *tree) rebuildHelper(pk interface{}, left, treeId, level int) (int, error) {
	right := left + 1
	var children []int
	emptyNode := reflectNew(t.node)
	// 以原有lft为序修正Tree
	err := t.Model(emptyNode).
		Select(t.colID()).Where(t.colParent()+" = ?", pk).Order(t.colLeft() + " ASC").Scan(&children).Error
	if err != nil {
		return 0, err
	}
	for _, child := range children {
		right, err = t.rebuildHelper(child, right, treeId, level+1)
		if err != nil {
			return right + 1, err
		}
	}
	err = t.Model(emptyNode).Where(t.colID()+" = ?", pk).
		Select(t.colTree(), t.colLeft(), t.colRight(), t.colLevel()).
		Updates(map[string]interface{}{
			t.colTree(true):  treeId,
			t.colLeft(true):  left,
			t.colRight(true): right,
			t.colLevel(true): level,
		}).Error
	return right + 1, err
}
