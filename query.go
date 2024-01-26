package mptt

import (
	"gorm.io/gorm"
	"math"
)

// RefreshNode 刷新节点信息，在Insert， Move， Delete操作之后，
// 有可能导致某个节点的信息已发生变化，需要通过id get到节点来刷新位置信息
func (t *tree) RefreshNode(n interface{}) error {
	return t.Table(t.tableName).Where(t.colID()+" = ?", t.getNodeID(n)).Find(n).Error
}

func (t *tree) getDescendantCount(n interface{}) int {
	return int(math.Floor(float64(t.getRight(n)-t.getLeft(n)-1) / 2))
}

func (t *tree) GetDescendantCount() int {
	return t.getDescendantCount(t.node)
}

func (t *tree) GetAncestors(outListPtr interface{}, ascending, includeSelf bool) error {
	var (
		leftCol = t.colLeft()
	)
	order := leftCol + " desc"
	if ascending {
		order = leftCol + " asc"
	}
	whereSql := "[tree_id] = ?"
	if includeSelf {
		whereSql += " AND [left] <= ? AND [right] >= ?"
	} else {
		whereSql += " AND [left] < ? AND [right] > ?"
	}
	return t.Model(reflectNew(t.node)).
		Where(t.replacePlaceholder(whereSql),
			t.getTreeID(t.node),
			t.getLeft(t.node),
			t.getRight(t.node),
		).Order(order).Find(outListPtr).Error
}

func (t *tree) GetDescendants(outListPtr interface{}, includeSelf bool) error {
	whereSql := "[tree_id] = ?"
	if includeSelf {
		whereSql += " AND [left] >= ? AND [right] <= ?"
	} else {
		whereSql += " AND [left] > ? AND [right] < ?"
	}
	return t.Model(reflectNew(t.node)).
		Where(t.replacePlaceholder(whereSql),
			t.getTreeID(t.node),
			t.getLeft(t.node),
			t.getRight(t.node),
		).Order(t.colLeft() + " asc").Find(outListPtr).Error
}

func (t *tree) GetFamily(outListPtr interface{}) error {
	var (
		treeId = t.getTreeID(t.node)
		left   = t.getLeft(t.node)
		right  = t.getRight(t.node)
	)
	whereSql := "[tree_id] = ? AND (([left] <= ? AND [right] >= ?) OR ( [left] > ? AND [right] < ? ))"
	return t.Where(
		t.replacePlaceholder(whereSql),
		treeId,
		left,
		right,
		left,
		right,
	).Order(t.colLeft() + " ASC").Find(outListPtr).Error
}

func (t *tree) GetChildren(outListPtr interface{}) error {
	whereSql := "[tree_id] = ? AND [parent_id] = ?"
	return t.Model(reflectNew(t.node)).
		Where(t.replacePlaceholder(whereSql),
			t.getTreeID(t.node),
			t.getNodeID(t.node),
		).
		Order(t.colLeft() + " asc").
		Find(outListPtr).Error
}

func (t *tree) GetLeafNodes(outListPtr interface{}) error {
	var (
		leftCol  = t.colLeft()
		rightCol = t.colRight()
		treeId   = t.getTreeID(t.node)
		left     = t.getLeft(t.node)
		right    = t.getRight(t.node)
	)
	whereSql := "[tree_id] = ? AND [left] > ? AND [right] < ?"
	return t.Model(reflectNew(t.node)).
		Where(t.replacePlaceholder(whereSql), treeId, left, right).
		Where(gorm.Expr(rightCol + " - " + leftCol + " = 1")).
		Order(t.colLeft() + " asc").
		Find(outListPtr).Error
}

func (t *tree) GetSiblings(outListPtr interface{}, includeSelf bool) error {
	tx := t.Model(reflectNew(t.node)).
		Where(t.colParent()+" = ?", t.getParentID(t.node))
	if !includeSelf {
		tx = tx.Where(t.colID()+" <> ?", t.getNodeID(t.node))
	}
	return tx.Order(t.colLeft() + " asc").Find(outListPtr).Error
}

func (t *tree) nextSibling(tx *gorm.DB, node interface{}) *gorm.DB {
	whereSql := t.replacePlaceholder("[parent_id] = ? AND [left] > ?")
	return tx.Where(whereSql, t.getParentID(node), t.getRight(node)).
		Order(t.colLeft() + " asc")
}

func (t *tree) GetNextSibling(outPtr interface{}, conds ...interface{}) error {
	if t.isRootNode(t.node) {
		return t.Where(t.colTree()+"> ?", t.getTreeID(t.node)).First(outPtr, conds...).Error
	}
	return t.Scopes(func(tx *gorm.DB) *gorm.DB {
		return t.nextSibling(tx, t.node)
	}).First(outPtr, conds...).Error
}

func (t *tree) previousSibling(tx *gorm.DB, node interface{}) *gorm.DB {
	whereSql := t.replacePlaceholder("[parent_id] = ? AND [right] < ?")
	return tx.Where(whereSql, t.getParentID(node), t.getLeft(node)).
		Order(t.colRight() + " desc")
}

func (t *tree) GetPreviousSibling(outPtr interface{}, conds ...interface{}) error {
	if t.isRootNode(t.node) {
		return t.Where(t.colTree()+"< ?", t.getTreeID(t.node)).First(outPtr, conds...).Error
	}
	return t.Scopes(func(tx *gorm.DB) *gorm.DB {
		return t.previousSibling(tx, t.node)
	}).First(outPtr, conds...).Error
}

func (t *tree) GetRoot(outPtr interface{}) error {
	return t.RootNode(t.getTreeID(t.node), outPtr)
}

func (t *tree) GetLevel() int {
	return t.getLevel(t.node)
}

func (t *tree) IsChildNode() bool {
	return t.isChildNode(t.node)
}

func (t *tree) isChildNode(n interface{}) bool {
	return !t.isRootNode(n)
}

func (t *tree) IsRootNode() bool {
	return t.isRootNode(t.node)
}

func (t *tree) isRootNode(n interface{}) bool {
	return isEmpty(t.getParentID(n))
}

func (t *tree) IsLeafNode() bool {
	return t.getDescendantCount(t.node) == 0
}

func (t *tree) IsDescendantOf(other interface{}, includeSelf bool) bool {
	var (
		otherID = t.getNodeID(other)
		curID   = t.getNodeID(t.node)
	)
	if includeSelf && t.equalIDValue(curID, otherID) {
		return true
	}
	if t.getTreeID(other) != t.getTreeID(t.node) {
		return false
	}
	return t.getLeft(t.node) > t.getLeft(other) && t.getRight(t.node) < t.getRight(other)
}

func (t *tree) IsAncestorOf(other interface{}, includeSelf bool) bool {
	var (
		otherID = t.getNodeID(other)
		curID   = t.getNodeID(t.node)
	)
	if includeSelf && t.equalIDValue(curID, otherID) {
		return true
	}
	if t.getTreeID(other) != t.getTreeID(t.node) {
		return false
	}
	return t.getLeft(t.node) < t.getLeft(other) && t.getRight(t.node) > t.getRight(other)
}

func (t *tree) RootNodes(outListPtr interface{}) error {
	emptyNode := reflectNew(t.node)
	return t.Model(emptyNode).Where(t.colParent()+" = ?", t.getParentID(emptyNode)).Find(outListPtr).Error
}

func (t *tree) RootNode(treeID int, outPtr interface{}) error {
	emptyNode := reflectNew(t.node)
	whereSql := t.replacePlaceholder("[parent_id] = ? AND [tree_id] = ?")
	return t.Model(emptyNode).Where(whereSql, t.getParentID(emptyNode), treeID).Find(outPtr).Error
}
