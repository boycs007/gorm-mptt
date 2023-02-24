package mptt

import (
    "errors"
    "fmt"
    "gorm.io/gorm"
    "math"
    "reflect"
)

const (
    defaultNewTreeId = -1
)

func (db *Tree) GetTableName(n interface{}) string {
    t := reflect.TypeOf(n)
    if t.Kind() == reflect.Ptr {
        t = t.Elem()
    }
    return db.NamingStrategy.TableName(t.Name())
}

func (db *Tree) MoveNode(n, targetPtr interface{}, position PositionEnum) (bool, error) {
    var err error
    if err = db.validateType(n); err != nil {
        return false, err
    }
    if err = db.validateType(targetPtr); err != nil {
        return false, err
    }
    if targetPtr == nil {
        if db.innerNode(n).IsChildNode() {
            err = db.makeChildRootNode(n, defaultNewTreeId)
        }
    } else if db.innerNode(targetPtr).IsRootNode() && (position == Left || position == Right) {
        err = db.makeSiblingOfRootNode(n, targetPtr, position)
    } else {
        if db.innerNode(n).IsRootNode() {
            err = db.moveRootNode(n, targetPtr, position)
        } else {
            err = db.moveChildNode(n, targetPtr, position)
        }
    }
    return err == nil, err
}

// make target node and it's descendants to a new tree
// and fix the left & right value in the old tree
func (db *Tree) makeChildRootNode(n interface{}, treeId int) error {
    var (
        node      = db.getModelBase(n)
        lvlOffset = node.Lvl
        gapSize   = node.Rght - node.Lft + 1
        offset    = node.Lft - 1
        updateSQL = `UPDATE ` + db.GetTableName(n) +
            ` SET lvl = CASE
                WHEN lft >= ? AND lft <= ?
                    THEN lvl - ?
                ELSE lvl END,
            tree_id = CASE
                WHEN lft >= ? AND lft <= ?
                    THEN ?
                ELSE tree_id END,
            lft = CASE
                WHEN lft >= ? AND lft <= ?
                    THEN lft - ?
                WHEN lft > ?
                    THEN lft - ?
                ELSE lft END,
            rght = CASE
                WHEN rght >= ? AND rght <= ?
                    THEN rght - ?
                WHEN rght > ?
                    THEN rght - ?
                ELSE rght END
        WHERE tree_id = ?`
    )
    if treeId <= 0 {
        treeId = db.getNextTreeId(n)
    }

    err := db.Statement.Exec(updateSQL,
        node.Lft, // new tree lvl param
        node.Rght,
        lvlOffset,

        node.Lft, // new tree tree_id param
        node.Rght,
        treeId,

        node.Lft, // new tree left param
        node.Rght,
        offset,

        node.Rght, // old tree left fix
        gapSize,

        node.Lft, // new tree right
        node.Rght,
        offset,

        node.Rght, // old tree right fix
        gapSize,

        node.TreeID, // where condition
    ).Error

    if err != nil {
        return err
    }
    // fix the values of target model in memory
    node.TreeID = treeId
    node.ParentID = 0
    node.Lvl = 0 // base.Lvl - lvlOffset
    node.Lft = 1 // base.Lft - offset
    node.Rght = node.Rght - offset
    db.setModelBase(n, &node)
    return db.Statement.DB.Model(n).Select("tree_id", "lvl", "lft", "rght", "parent_id").Updates(n).Error
}

func (db *Tree) makeSiblingOfRootNode(n, targetPtr interface{}, position PositionEnum) error {
    var (
        err                    error
        spaceTarget, newTreeId int
        targetNode             = db.getModelBase(targetPtr)
        node                   = db.getModelBase(n)
    )
    if db.innerNode(n).IsChildNode() {
        if position == Left {
            spaceTarget = targetNode.TreeID - 1
            newTreeId = targetNode.TreeID
        } else if position == Right {
            spaceTarget = targetNode.TreeID
            newTreeId = targetNode.TreeID + 1
        } else {
            return fmt.Errorf("an invalid position was given: %s", position)
        }
        err = db.createTreeSpace(n, spaceTarget, 1)
        if err != nil {
            return err
        }
        if node.TreeID > spaceTarget {
            // node.TreeID has been incremented by createTreeSpace in the database
            node.TreeID = node.TreeID + 1
            db.setModelBase(n, &node)
        }
        return db.makeChildRootNode(n, newTreeId)
    }

    var (
        shift, lowerBound, upperBound int
        leftSibling, rightSibling     MPTTModelBase
    )
    // not childNode
    if position == Left {
        if targetNode.TreeID > node.TreeID {
            err = db.Statement.DB.Model(targetPtr).Select("id", "tree_id").Scopes(func(tx *gorm.DB) *gorm.DB {
                return db.previousSibling(tx, &targetNode)
            }).First(&leftSibling).Error
            if err != nil {
                return err
            }
            if node.ID == leftSibling.ID {
                return nil
            }
            newTreeId = leftSibling.TreeID
            lowerBound = node.TreeID
            upperBound = newTreeId
            shift = -1
        } else {
            newTreeId = targetNode.TreeID
            lowerBound = newTreeId
            upperBound = node.TreeID
            shift = 1
        }
    } else if position == Right {
        if targetNode.TreeID > node.TreeID {
            newTreeId = targetNode.TreeID
            lowerBound = node.TreeID
            upperBound = newTreeId
            shift = -1
        } else {
            err = db.Statement.DB.Model(targetPtr).Select("id", "tree_id").Scopes(func(tx *gorm.DB) *gorm.DB {
                return db.nextSibling(tx, &targetNode)
            }).First(&rightSibling).Error
            if err != nil {
                return err
            }
            if node.ID == rightSibling.ID {
                return nil
            }
            newTreeId = rightSibling.TreeID
            lowerBound = newTreeId
            upperBound = node.TreeID
            shift = 1
        }
    } else {
        return fmt.Errorf("an invalid position was given: %s", position)
    }
    rootSiblingUpdateSql := `UPDATE ` + db.GetTableName(n) +
        ` SET tree_id = CASE
                WHEN tree_id = ?
                THEN ?
                ELSE tree_id + ? END
    WHERE tree_id >= ? AND tree_id <= ?`

    err = db.Statement.DB.Exec(rootSiblingUpdateSql,
        node.TreeID,
        newTreeId,
        shift,
        lowerBound,
        upperBound,
    ).Error
    if err != nil {
        return err
    }
    node.TreeID = newTreeId
    db.setModelBase(n, &node)
    return db.Statement.DB.Model(n).Select("tree_id").Updates(n).Error
}

func (db *Tree) moveChildNode(n, targetPtr interface{}, position PositionEnum) error {
    var (
        targetNode = db.getModelBase(targetPtr)
        node       = db.getModelBase(n)
    )
    if node.TreeID == targetNode.TreeID {
        return db.moveChildWithinTree(n, targetPtr, position)
    }
    return db.moveChildToNewTree(n, targetPtr, position)
}

func (db *Tree) moveChildWithinTree(n, targetPtr interface{}, position PositionEnum) error {
    var (
        err               error
        targetNode        = db.getModelBase(targetPtr)
        node              = db.getModelBase(n)
        width             = node.Rght - node.Lft + 1
        newLeft, newRight int
        lvlOffset         int
        parentId          int
    )
    if position == LastChild || position == FirstChild {
        if node.ID == targetNode.ID {
            return errors.New("a node may not be made a child of itself")
        } else if node.Lft < targetNode.Lft && targetNode.Lft < node.Rght {
            return errors.New("a node may not be made a child of any of its descendants")
        }
        if position == LastChild {
            if targetNode.Rght > node.Rght {
                newLeft = targetNode.Rght - width
                newRight = targetNode.Rght - 1
            } else {
                newLeft = targetNode.Rght
                newRight = targetNode.Rght + width - 1
            }
        } else {
            if targetNode.Lft > node.Lft {
                newLeft = targetNode.Lft - width + 1
                newRight = targetNode.Lft
            } else {
                newLeft = targetNode.Lft + 1
                newRight = targetNode.Lft + width
            }
        }
        lvlOffset = node.Lvl - targetNode.Lvl - 1
        parentId = targetNode.ID
    } else if position == Left || position == Right {
        if node.ID == targetNode.ID {
            return errors.New("a node may not be made a sibling of itself")
        } else if node.Lft < targetNode.Lft && targetNode.Lft < node.Rght {
            return errors.New("a node may not be made a sibling of any of its descendants")
        }
        if position == Left {
            if targetNode.Lft > node.Lft {
                newLeft = targetNode.Lft - width
                newRight = targetNode.Lft - 1
            } else {
                newLeft = targetNode.Lft
                newRight = targetNode.Lft + width - 1
            }
        } else {
            if targetNode.Rght > node.Rght {
                newLeft = targetNode.Rght - width + 1
                newRight = targetNode.Rght
            } else {
                newLeft = targetNode.Rght + 1
                newRight = targetNode.Rght + width
            }
        }
        lvlOffset = node.Lvl - targetNode.Lvl
        parentId = targetNode.ParentID
    } else {
        return fmt.Errorf("an invalid position was given: %s", position)
    }

    leftBoundary := int(math.Min(float64(node.Lft), float64(newLeft)))
    rightBoundary := int(math.Max(float64(node.Rght), float64(newRight)))
    offset := newLeft - node.Lft
    gapSize := width
    if offset > 0 {
        gapSize = -gapSize
    }

    moveSubtreeSql := `UPDATE ` + db.GetTableName(n) +
        ` SET lvl = CASE
                WHEN lft >= ? AND lft <= ?
                    THEN lvl - ?
                ELSE lvl END,
            lft = CASE
                WHEN lft >= ? AND lft <= ?
                    THEN lft + ?
                WHEN lft >= ? AND lft <= ?
                    THEN lft + ?
                ELSE lft END,
            rght = CASE
                WHEN rght >= ? AND rght <= ?
                    THEN rght + ?
                WHEN rght >= ? AND rght <= ?
                    THEN rght + ?
                ELSE rght END
    WHERE tree_id = ?`

    err = db.Statement.DB.Exec(moveSubtreeSql,
        node.Lft,
        node.Rght,
        lvlOffset,

        node.Lft,
        node.Rght,
        offset,

        leftBoundary,
        rightBoundary,
        gapSize,

        node.Lft,
        node.Rght,
        offset,

        leftBoundary,
        rightBoundary,
        gapSize,

        node.TreeID,
    ).Error
    if err != nil {
        return err
    }
    node.Lft = newLeft
    node.Rght = newRight
    node.Lvl = node.Lvl - lvlOffset
    node.ParentID = parentId
    db.setModelBase(n, &node)
    return db.Statement.DB.Model(n).Select("lft", "rght", "lvl", "parent_id").Updates(n).Error
}

func (db *Tree) moveChildToNewTree(n, targetPtr interface{}, position PositionEnum) error {
    var (
        err                                      error
        targetNode                               = db.getModelBase(targetPtr)
        node                                     = db.getModelBase(n)
        treeWidth                                int
        spaceTarget, lvlOffset, offset, parentId int
    )
    spaceTarget, lvlOffset, offset, parentId, _, err = db.calculateInterTreeMoveValues(n, targetPtr, position)
    if err != nil {
        return err
    }
    treeWidth = node.Lft - node.Rght + 1

    if err = db.manageSpace(n, treeWidth, spaceTarget, targetNode.TreeID); err != nil {
        return err
    }

    if err = db.interTreeMoveAndCloseGap(node, lvlOffset, offset, targetNode.TreeID); err != nil {
        return err
    }
    node.Lft = node.Lft - offset
    node.Rght = node.Rght - offset
    node.Lvl = node.Lvl - lvlOffset
    node.TreeID = targetNode.TreeID
    node.ParentID = parentId
    db.setModelBase(n, &node)
    return db.Statement.DB.Model(n).Select("tree_id", "lft", "rght", "lvl", "parent_id").Updates(n).Error
}

func (db *Tree) moveRootNode(n, targetPtr interface{}, position PositionEnum) error {
    var (
        err                                      error
        targetNode                               = db.getModelBase(targetPtr)
        node                                     = db.getModelBase(n)
        width                                    = node.Rght - node.Lft + 1
        spaceTarget, lvlOffset, offset, parentId int
    )
    if node.ID == targetNode.ID {
        return errors.New("a node may not be made a child of itself")
    } else if node.TreeID == targetNode.TreeID {
        return errors.New("a node may not be made a child of any of its descendants")
    }

    spaceTarget, lvlOffset, offset, parentId, _, err = db.calculateInterTreeMoveValues(n, targetPtr, position)
    if err != nil {
        return err
    }
    if err = db.manageSpace(n, width, spaceTarget, targetNode.TreeID); err != nil {
        return err
    }

    moveTreeSql := `UPDATE ` + db.GetTableName(n) +
        ` SET lvl = lvl - ?,
        lft = lft - ?,
        rght = rght - ?,
        tree_id = ?
        WHERE lft >= ? AND lft <= ?
        AND tree_id = ?`

    if err = db.Exec(moveTreeSql,
        lvlOffset,
        offset,
        offset,
        targetNode.TreeID,
        node.Lft,
        node.Rght,
        node.TreeID,
    ).Error; err != nil {
        return err
    }

    node.Lft = node.Lft - offset
    node.Rght = node.Rght - offset
    node.Lvl = node.Lvl - lvlOffset
    node.TreeID = targetNode.TreeID
    node.ParentID = parentId
    db.setModelBase(n, &node)
    return db.Statement.DB.Model(n).Select("tree_id", "lft", "rght", "lvl", "parent_id").Updates(n).Error
}

// 用于在删除掉元素后，将树上的左右修正
func (db *Tree) closeGap(n interface{}, size, target, treeId int) error {
    return db.manageSpace(n, -size, target, treeId)
}

// 用于在插入了新元素后，将树上的左右修正
func (db *Tree) createSpace(n interface{}, size, target, treeId int) error {
    return db.manageSpace(n, size, target, treeId)
}

// 根据target来修改特定tree_id的树上的lft rght值
// lft > target的，修改为 lft + size， 否则不修改
// rght > target的，修改为 rght + size, 否则不修改
func (db *Tree) manageSpace(n interface{}, size, target, treeId int) error {
    spaceSql := `UPDATE ` + db.GetTableName(n) +
        ` SET lft = CASE
                WHEN lft > ?
                    THEN lft + ?
                ELSE lft END,
            rght = CASE
                WHEN rght > ?
                    THEN rght + ?
                ELSE rght END
        WHERE tree_id = ? AND (lft > ? OR rght > ?)`

    return db.Exec(spaceSql,
        target,
        size,
        target,
        size,
        treeId,
        target,
        target,
    ).Error
}

func (db *Tree) interTreeMoveAndCloseGap(n MPTTModelBase, lvlOffset, offset, newTreeId int) error {
    sql := `UPDATE ` + db.GetTableName(n) +
        ` SET lvl = CASE
                WHEN lft >= ? AND lft <= ?
                    THEN lvl - ?
                ELSE lvl END,
            tree_id = CASE
                WHEN lft >= ? AND lft <= ?
                    THEN ?
                ELSE tree_id END,
            lft = CASE
                WHEN lft >= ? AND lft <= ?
                    THEN lft - ?
                WHEN lft > ?
                    THEN lft - ?
                ELSE lft END,
            rght = CASE
                WHEN rght >= ? AND rght <= ?
                    THEN rght - ?
                WHEN rght > ?
                    THEN rght - ?
                ELSE rght END
        WHERE tree_id = ?`

    gapSize := n.Rght - n.Lft + 1
    gapTargetLeft := n.Lft - 1
    return db.Exec(sql,
        n.Lft,
        n.Rght,
        lvlOffset,
        n.Lft,
        n.Rght,
        newTreeId,
        n.Lft,
        n.Rght,
        offset,
        gapTargetLeft,
        gapSize,
        n.Lft,
        n.Rght,
        offset,
        gapTargetLeft,
        gapSize,
        n.TreeID,
    ).Error
}

func (db *Tree) calculateInterTreeMoveValues(n, targetPtr interface{}, position PositionEnum) (
    spaceTarget,
    lvlOffset,
    offset,
    parentId,
    rightShift int,
    err error,
) {
    var (
        targetNode = db.getModelBase(targetPtr)
        node       = db.getModelBase(n)
    )

    if position == LastChild || position == FirstChild {
        if position == LastChild {
            spaceTarget = targetNode.Rght - 1
        } else {
            spaceTarget = targetNode.Lft
        }
        lvlOffset = node.Lvl - targetNode.Lvl - 1
        parentId = targetNode.ID
    } else if position == Left || position == Right {
        if position == Left {
            spaceTarget = targetNode.Lft - 1
        } else {
            spaceTarget = targetNode.Rght
        }
        lvlOffset = node.Lvl - targetNode.Lvl
        parentId = targetNode.ParentID
    } else {
        err = fmt.Errorf("an invalid position was given: %s", position)
    }

    offset = node.Lft - spaceTarget - 1

    rightShift = 0
    if parentId > 0 {
        rightShift = 2 * (db.innerNode(n).GetDescendantCount() + 1)
    }
    return
}
