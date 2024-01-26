package mptt

import (
	"errors"
	"fmt"
	"math"
)

const (
	defaultNewTreeId = -1
)

func (t *tree) MoveNodeByID(nodeID, targetID interface{}, position PositionEnum) (bool, error) {
	n, err := t.getNodeByID(nodeID)
	if err != nil {
		return false, err
	}
	target, err := t.getNodeByID(targetID)
	if err != nil {
		return false, err
	}
	return t.MoveNode(n, target, position)
}

func (t *tree) MoveNode(n, targetPtr interface{}, position PositionEnum, refreshTarget ...bool) (bool, error) {
	var err error
	if err = t.validateType(n); err != nil {
		return false, err
	}
	if targetPtr != nil {
		if err = t.validateType(targetPtr); err != nil {
			return false, err
		}
	}
	if targetPtr == nil {
		if t.isChildNode(n) {
			err = t.makeChildRootNode(n, defaultNewTreeId)
		}
	} else if t.isRootNode(targetPtr) && (position == Left || position == Right) {
		err = t.makeSiblingOfRootNode(n, targetPtr, position)
	} else {
		if t.isRootNode(n) {
			err = t.moveRootNode(n, targetPtr, position)
		} else {
			err = t.moveChildNode(n, targetPtr, position)
		}
	}
	if len(refreshTarget) > 0 && refreshTarget[0] {
		err = t.Model(reflectNew(n)).First(targetPtr).Error
	}
	return err == nil, err
}

// make target node and it's descendants to a new tree
// and fix the left & right value in the old tree
func (t *tree) makeChildRootNode(n interface{}, treeId int) error {

	var (
		err         error
		rawTreeID   = t.getTreeID(n)
		lvl         = t.getLevel(n)
		lft         = t.getLeft(n)
		rght        = t.getRight(n)
		lvlOffset   = lvl - 1
		gapSize     = rght - lft + 1
		offset      = lft - 1
		colLft      = t.colLeft()
		colRght     = t.colRight()
		colTreeId   = t.colTree()
		colLvl      = t.colLevel()
		colParentId = t.colParent()
		updateSQL   = t.replacePlaceholder(`UPDATE [table_tree] SET 
        [level] = CASE WHEN [left] >= ? AND [left] <= ? THEN [level] - ? ELSE [level] END,
        [tree_id] = CASE WHEN [left] >= ? AND [left] <= ? THEN ? ELSE [tree_id] END,
        [left] = CASE WHEN [left] >= ? AND [left] <= ? THEN [left] - ? WHEN [left] > ? THEN [left] - ? ELSE [left] END,
        [right] = CASE WHEN [right] >= ? AND [right] <= ? THEN [right] - ? WHEN [right] > ? THEN [right] - ? ELSE [right] END
        WHERE [tree_id] = ?`)
	)
	if treeId <= 0 {
		treeId = t.getNextTreeId()
	}

	err = t.Statement.Exec(updateSQL,
		lft, // new tree colLvl param
		rght,
		lvlOffset,

		lft, // new tree tree_id param
		rght,
		treeId,

		lft, // new tree left param
		rght,
		offset,

		rght, // old tree left fix
		gapSize,

		lft, // new tree right
		rght,
		offset,

		rght, // old tree right fix
		gapSize,

		rawTreeID, // where condition
	).Error

	if err != nil {
		return err
	}
	// fix the values of target model in memory
	emptyNode := reflectNew(t.node)
	t.setTreeID(n, treeId)
	t.setParentID(n, t.getParentID(emptyNode))
	t.setLevel(n, 1)
	t.setLeft(n, 1)
	t.setRight(n, rght-offset)
	return t.Model(emptyNode).
		Select(
			colTreeId,
			colLvl,
			colLft,
			colRght,
			colParentId,
		).Where(t.colID()+" = ?", t.getNodeID(n)).Updates(n).Error
}

func (t *tree) moveChildNode(n, targetPtr interface{}, position PositionEnum) error {
	var (
		nTreeId      = t.getTreeID(n)
		targetTreeId = t.getTreeID(targetPtr)
	)
	if nTreeId == targetTreeId {
		return t.moveChildWithinTree(n, targetPtr, position)
	}
	return t.moveChildToNewTree(n, targetPtr, position)
}

func (t *tree) moveChildWithinTree(n, targetPtr interface{}, position PositionEnum) error {
	var (
		err               error
		tRght             = t.getRight(targetPtr)
		tLft              = t.getLeft(targetPtr)
		tLvl              = t.getLevel(targetPtr)
		tid               = t.getNodeID(targetPtr)
		tPid              = t.getParentID(targetPtr)
		lft               = t.getLeft(n)
		rght              = t.getRight(n)
		lvl               = t.getLevel(n)
		id                = t.getNodeID(n)
		treeId            = t.getTreeID(n)
		width             = rght - lft + 1
		newLeft, newRight int
		lvlOffset         int
		parentID          interface{}
	)
	if position == LastChild || position == FirstChild {
		if t.equalIDValue(id, tid) {
			return errors.New("a node may not be made a child of itself")
		} else if lft < tLft && tLft < rght {
			return errors.New("a node may not be made a child of any of its descendants")
		}
		if position == LastChild {
			if tRght > rght {
				newLeft = tRght - width
				newRight = tRght - 1
			} else {
				newLeft = tRght
				newRight = tRght + width - 1
			}
		} else {
			if tLft > lft {
				newLeft = tLft - width + 1
				newRight = tLft
			} else {
				newLeft = tLft + 1
				newRight = tLft + width
			}
		}
		lvlOffset = lvl - tLvl - 1
		parentID = tid
	} else if position == Left || position == Right {
		if t.equalIDValue(id, tid) {
			return errors.New("a node may not be made a sibling of itself")
		} else if lft < tLft && tLft < rght {
			return errors.New("a node may not be made a sibling of any of its descendants")
		}
		if position == Left {
			if tLft > lft {
				newLeft = tLft - width
				newRight = tLft - 1
			} else {
				newLeft = tLft
				newRight = tLft + width - 1
			}
		} else {
			if tRght > rght {
				newLeft = tRght - width + 1
				newRight = tRght
			} else {
				newLeft = tRght + 1
				newRight = tRght + width
			}
		}
		lvlOffset = lvl - tLvl
		parentID = tPid
	} else {
		return fmt.Errorf("an invalid position was given: %s", position)
	}

	leftBoundary := int(math.Min(float64(lft), float64(newLeft)))
	rightBoundary := int(math.Max(float64(rght), float64(newRight)))
	offset := newLeft - lft
	gapSize := width
	if offset > 0 {
		gapSize = -gapSize
	}

	moveSubtreeSql := t.replacePlaceholder(`UPDATE [table_tree] SET 
		[level] = CASE WHEN [left] >= ? AND [left] <= ? THEN [level] - ? ELSE [level] END,
        [left] = CASE WHEN [left] >= ? AND [left] <= ? THEN [left] + ? WHEN [left] >= ? AND [left] <= ? THEN [left] + ? ELSE [left] END,
        [right] = CASE WHEN [right] >= ? AND [right] <= ? THEN [right] + ? WHEN [right] >= ? AND [right] <= ? THEN [right] + ? ELSE [right] END
    	WHERE [tree_id] = ?`)

	err = t.Exec(moveSubtreeSql,
		lft,
		rght,
		lvlOffset,

		lft,
		rght,
		offset,

		leftBoundary,
		rightBoundary,
		gapSize,

		lft,
		rght,
		offset,

		leftBoundary,
		rightBoundary,
		gapSize,

		treeId,
	).Error
	if err != nil {
		return err
	}
	newLvl := lvl - lvlOffset
	t.setLeft(n, newLeft)
	t.setRight(n, newRight)
	t.setLevel(n, newLvl)
	t.setParentID(n, parentID)

	return t.Statement.DB.Model(reflectNew(n)).Select(
		t.colLeft(),
		t.colRight(),
		t.colLevel(),
		t.colParent(),
	).Where(t.colID()+"= ?", t.getNodeID(n)).Updates(n).Error
}

func (t *tree) interTreeMoveAndCloseGap(n interface{}, lvlOffset, offset, newTreeId int) error {
	var (
		err    error
		lft    = t.getLeft(n)
		rght   = t.getRight(n)
		treeId = t.getTreeID(n)
	)
	sql := t.replacePlaceholder(`UPDATE [table_tree] SET 
		[level] = CASE WHEN [left] >= ? AND [left] <= ? THEN [level] - ? ELSE [level] END,
        [tree_id] = CASE WHEN [left] >= ? AND [left] <= ? THEN ? ELSE [tree_id] END,
        [left] = CASE WHEN [left] >= ? AND [left] <= ? THEN [left] - ? WHEN [left] > ? THEN [left] - ? ELSE [left] END,
        [right] = CASE WHEN [right] >= ? AND [right] <= ? THEN [right] - ? WHEN [right] > ? THEN [right] - ? ELSE [right] END
        WHERE [tree_id] = ?`)

	gapSize := rght - lft + 1
	gapTargetLeft := lft - 1
	err = t.Exec(sql,
		lft,
		rght,
		lvlOffset,
		lft,
		rght,
		newTreeId,
		lft,
		rght,
		offset,
		gapTargetLeft,
		gapSize,
		lft,
		rght,
		offset,
		gapTargetLeft,
		gapSize,
		treeId,
	).Error
	return err
}

func (t *tree) calculateInterTreeMoveValues(n, targetPtr interface{}, position PositionEnum) (
	spaceTarget,
	lvlOffset,
	offset,
	rightShift int,
	parentId interface{},
	err error,
) {
	var (
		tRght = t.getRight(targetPtr)
		tLft  = t.getLeft(targetPtr)
		tLvl  = t.getLevel(targetPtr)
		tId   = t.getNodeID(targetPtr)
		tPid  = t.getParentID(targetPtr)
		lft   = t.getLeft(n)
		lvl   = t.getLevel(n)
	)

	if position == LastChild || position == FirstChild {
		if position == LastChild {
			spaceTarget = int(tRght) - 1
		} else {
			spaceTarget = tLft
		}
		lvlOffset = lvl - tLvl - 1
		parentId = tId
	} else if position == Left || position == Right {
		if position == Left {
			spaceTarget = int(tLft) - 1
		} else {
			spaceTarget = tRght
		}
		lvlOffset = lvl - tLvl
		parentId = tPid
	} else {
		err = fmt.Errorf("an invalid position was given: %s", position)
	}

	offset = lft - spaceTarget - 1

	rightShift = 0
	if !isEmpty(parentId) {
		rightShift = 2 * (int(t.getDescendantCount(n)) + 1)
	}
	return
}

func (t *tree) moveRootNode(n, targetPtr interface{}, position PositionEnum) error {
	var (
		err     error
		tid     = t.getNodeID(targetPtr)
		tTreeId = t.getTreeID(targetPtr)
		lft     = t.getLeft(n)
		rght    = t.getRight(n)
		lvl     = t.getLevel(n)
		id      = t.getNodeID(n)
		treeId  = t.getTreeID(n)
		width   = rght - lft + 1

		spaceTarget, lvlOffset, offset int
		parentId                       interface{}
	)
	if t.equalIDValue(id, tid) {
		return errors.New("a node may not be made a child/sibling of itself")
	} else if treeId == tTreeId {
		return errors.New("a node may not be made a child of any of its descendants")
	}

	spaceTarget, lvlOffset, offset, _, parentId, err = t.calculateInterTreeMoveValues(n, targetPtr, position)
	if err != nil {
		return err
	}
	if err = t.manageSpace(width, spaceTarget, tTreeId); err != nil {
		return err
	}

	moveTreeSql := t.replacePlaceholder(`UPDATE [table_tree] SET 
		[level] = [level] - ?,
        [left] = [left] - ?,
        [right] = [right] - ?,
        [tree_id] = ?
        WHERE [left] >= ? AND [left] <= ?
        AND [tree_id] = ?`)

	err = t.Exec(moveTreeSql,
		lvlOffset,
		offset,
		offset,
		tTreeId,
		lft,
		rght,
		treeId,
	).Error
	if err != nil {
		return err
	}
	newLft := lft - offset
	newRght := rght - offset
	newLvl := lvl - lvlOffset
	newPid := parentId
	t.setLeft(n, newLft)
	t.setRight(n, newRght)
	t.setLevel(n, newLvl)
	t.setParentID(n, newPid)
	t.setTreeID(n, tTreeId)
	return t.Model(reflectNew(n)).Select(
		t.colTree(),
		t.colLeft(),
		t.colRight(),
		t.colLevel(),
		t.colParent(),
	).Where(t.colID()+" = ?", t.getNodeID(n)).Updates(n).Error
}

func (t *tree) makeSiblingOfRootNode(n, targetPtr interface{}, position PositionEnum) error {
	var (
		err     error
		tTreeId = t.getTreeID(targetPtr)
		id      = t.getNodeID(n)
		treeId  = t.getTreeID(n)

		spaceTarget, newTreeId int
	)
	if !t.isRootNode(n) {
		if position == Left {
			spaceTarget = tTreeId - 1
			newTreeId = tTreeId
		} else if position == Right {
			spaceTarget = tTreeId
			newTreeId = tTreeId + 1
		} else {
			return fmt.Errorf("an invalid position was given: %s", position)
		}
		err = t.createTreeSpace(n, spaceTarget, 1)
		if err != nil {
			return err
		}
		if treeId > spaceTarget {
			// node.TreeID has been incremented by createTreeSpace in the database
			t.setTreeID(n, treeId+1)
		}
		return t.makeChildRootNode(n, newTreeId)
	}

	var (
		shift, lowerBound, upperBound int
	)
	emptyNode := reflectNew(t.node)
	sibling := reflectNew(t.node)
	// not childNode
	if position == Left {
		if tTreeId > treeId {
			err = t.Model(emptyNode).Select(
				t.colID(),
				t.colTree(),
			).Where(t.colParent()+" = ? AND "+t.colTree()+" < ?",
				t.getParentID(emptyNode),
				tTreeId,
			).Order(t.colTree() + " desc").First(sibling).Error
			if err != nil {
				return err
			}
			if t.equalIDValue(id, t.getNodeID(sibling)) {
				return nil
			}
			newTreeId = t.getTreeID(sibling)
			lowerBound = treeId
			upperBound = newTreeId
			shift = -1
		} else {
			newTreeId = tTreeId
			lowerBound = newTreeId
			upperBound = treeId
			shift = 1
		}
	} else if position == Right {
		if tTreeId > treeId {
			newTreeId = tTreeId
			lowerBound = treeId
			upperBound = newTreeId
			shift = -1
		} else {
			err = t.Model(emptyNode).Select(
				t.colID(),
				t.colTree(),
			).Where(t.colParent()+" = ? AND "+t.colTree()+" > ?",
				t.getParentID(emptyNode),
				tTreeId,
			).Order(t.colTree() + " asc").
				First(sibling).Error
			if err != nil {
				return err
			}
			if t.equalIDValue(id, t.getNodeID(sibling)) {
				return nil
			}
			newTreeId = t.getTreeID(sibling)
			lowerBound = newTreeId
			upperBound = treeId
			shift = 1
		}
	} else {
		return fmt.Errorf("an invalid position was given: %s", position)
	}
	rootSiblingUpdateSql := t.replacePlaceholder(`UPDATE [table_tree] SET 
		[tree_id] = CASE WHEN [tree_id] = ? THEN ? ELSE [tree_id] + ? END
    	WHERE [tree_id] >= ? AND [tree_id] <= ?`)

	err = t.Exec(rootSiblingUpdateSql,
		treeId,
		newTreeId,
		shift,
		lowerBound,
		upperBound,
	).Error
	if err != nil {
		return err
	}
	t.setTreeID(n, newTreeId)
	return t.Model(emptyNode).Select(t.colTree()).Where(t.colID()+" = ?", t.getNodeID(n)).Updates(n).Error
}

func (t *tree) moveChildToNewTree(n, targetPtr interface{}, position PositionEnum) error {
	var (
		err      error
		parentId interface{}
		tTreeId  = t.getTreeID(targetPtr)
		lft      = t.getLeft(n)
		rght     = t.getRight(n)
		lvl      = t.getLevel(n)

		treeWidth, spaceTarget, lvlOffset, offset int
	)
	spaceTarget, lvlOffset, offset, _, parentId, err = t.calculateInterTreeMoveValues(n, targetPtr, position)
	if err != nil {
		return err
	}
	treeWidth = rght - lft + 1

	if err = t.manageSpace(treeWidth, spaceTarget, tTreeId); err != nil {
		return err
	}

	if err = t.interTreeMoveAndCloseGap(n, lvlOffset, offset, int(tTreeId)); err != nil {
		return err
	}
	newLft := lft - offset
	newRght := rght - offset
	newLvl := lvl - lvlOffset
	newPid := parentId
	t.setLeft(n, newLft)
	t.setRight(n, newRght)
	t.setLevel(n, newLvl)
	t.setParentID(n, newPid)
	t.setTreeID(n, tTreeId)
	return t.Table(t.tableName).Select(
		t.colTree(),
		t.colLeft(),
		t.colRight(),
		t.colLevel(),
		t.colTree(),
		t.colParent(),
	).Where(t.colID()+" = ?", t.getNodeID(n)).Updates(n).Error
}

// 用于在删除掉元素后，将树上的左右修正
func (t *tree) closeGap(size, target, treeId int) error {
	return t.manageSpace(-size, target, treeId)
}

// 用于在插入了新元素后，将树上的左右修正
func (t *tree) createSpace(size, target, treeId int) error {
	return t.manageSpace(size, target, treeId)
}

// 根据target来修改特定tree_id的树上的lft rght值
// lft > target的，修改为 lft + size， 否则不修改
// rght > target的，修改为 rght + size, 否则不修改
func (t *tree) manageSpace(size, target, treeId int) error {
	spaceSql := t.replacePlaceholder(`UPDATE [table_tree] SET 
[left] = CASE WHEN [left] > ? THEN [left] + ? ELSE [left] END,
[right] = CASE WHEN [right] > ? THEN [right] + ? ELSE [right] END
WHERE [tree_id] = ? AND ([left] > ? OR [right] > ?)`)

	err := t.Exec(spaceSql,
		target,
		size,
		target,
		size,
		treeId,
		target,
		target,
	).Error
	return err
}
