package mptt

import (
	"gorm.io/gorm"
)

func (t *tree) DeleteNodeByID(nodeID interface{}) error {
	// 查询一下确保数据是准的
	node := reflectNew(t.node)
	t.setNodeID(node, nodeID)
	err := t.Model(node).First(node).Error
	if err != nil {
		return err
	}
	return t.DeleteNode(node, true)
}

// DeleteNode delete current node and all descendants
func (t *tree) DeleteNode(n interface{}, doNotRefresh ...bool) error {
	var (
		err      error
		realNode = n
	)
	if len(doNotRefresh) == 0 || !doNotRefresh[0] {
		if err = t.validateType(n); err != nil {
			return err
		}
		realNode, err = t.getNodeByID(t.getNodeID(n))
		if err != nil {
			return err
		}
	}

	var (
		right      = t.getRight(realNode)
		left       = t.getLeft(realNode)
		parentID   = t.getParentID(realNode)
		treeID     = t.getTreeID(realNode)
		treeDbName = t.colTree()
	)
	diff := right - left + 1
	whereSql := t.replacePlaceholder("[tree_id] = ? AND [left] >= ? AND [left] < ?")
	emptyNode := reflectNew(realNode)
	err = t.Model(emptyNode).
		Where(whereSql,
			treeID, left, right).
		Delete(map[string]interface{}{}).Error
	if err != nil {
		return err
	}
	if isEmpty(parentID) {
		// delete the whole tree, close the tree id gap.
		return t.Model(emptyNode).
			Where(treeDbName+" > ?", treeID).
			Update(t.colTree(true), gorm.Expr(treeDbName+" - 1")).Error
	}
	return t.closeGap(diff, right, treeID)
}
