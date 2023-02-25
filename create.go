package mptt

// CreateNode 插入新节点。当节点的ParentID为0时，将生成新的树的根节点；
// 当节点的ParentID不为0时，将新节点插入为Parent的最后一个子节点
func (db *Tree) CreateNode(n interface{}) error {
    if err := db.validateType(n); err != nil {
        return err
    }
    ctx := db.getContext(n)
    base := ctx.ModelBase
    parentId := base.ParentID

    if parentId == 0 {
        // new tree root node
        base.TreeID = db.getNextTreeId(ctx.Node)
        base.Lft = 1
        base.Rght = 2
        base.Lvl = 1
        db.setModelBase(n, base)
        return db.Statement.Create(n).Error
    }
    parent := db.getNodeByParentId(ctx)
    return db.InsertNode(n, &parent, LastChild, false)
}

// InsertNode 插入新节点
//
// @param n: 需要插入的节点
// @param toPtr: 需要插入到哪里，toPtr为已存在的某个节点
// @param position:  需要插入到相对于toPtr的某个位置，可以是其
//   LastChild: 作为toPtr的最后一个子节点
//   FirstChild: 作为toPtr的第一个子节点
//   Left: 插入到toPtr的左边(前面)
//   Right: 插入到toPtr的右边(后面)
// @param refreshToPtr: 是否需要将toPtr对象的信息进行更新，例如如果插入到toPtr的左侧后，toPtr的lft、rght值将会更新
func (db *Tree) InsertNode(n, toPtr interface{}, position PositionEnum, refreshToPtr bool) error {
    var (
        edge         int
        parent, base *MPTTModelBase
    )
    if err := db.validateType(n); err != nil {
        return err
    }
    if err := db.validateType(toPtr); err != nil {
        return err
    }
    base = db.getContext(n).ModelBase
    existCtx := db.getContext(toPtr)
    base.TreeID = existCtx.ModelBase.TreeID

    // 根节点插入
    if existCtx.ModelBase.ParentID == 0 && (position == Left || position == Right) {
        var spaceTarget int
        if position == Left {
            base.TreeID = existCtx.ModelBase.TreeID
            spaceTarget = existCtx.ModelBase.TreeID - 1
            existCtx.ModelBase.TreeID = existCtx.ModelBase.TreeID + 1
        } else {
            base.TreeID = existCtx.ModelBase.TreeID + 1
            spaceTarget = existCtx.ModelBase.TreeID
        }
        base.Lft = 1
        base.Rght = 2
        base.Lvl = 1
        if err := db.createTreeSpace(n, spaceTarget, 1); err != nil {
            return err
        }
        db.setModelBase(n, base)
        if refreshToPtr {
            db.setModelBase(toPtr, existCtx.ModelBase)
        }
        return db.Statement.Create(n).Error
    }

    switch position {
    case LastChild:
        parent = existCtx.ModelBase
        edge = parent.Rght
        base.Lvl = parent.Lvl + 1
        base.ParentID = parent.ID
        if refreshToPtr {
            parent.Rght = parent.Rght + 2
        }
    case FirstChild:
        parent = existCtx.ModelBase
        edge = parent.Lft + 1
        base.Lvl = parent.Lvl + 1
        base.ParentID = parent.ID
        if refreshToPtr {
            parent.Rght = parent.Rght + 2
        }
    case Left:
        sibling := existCtx.ModelBase
        edge = sibling.Lft
        base.Lvl = sibling.Lvl
        base.ParentID = sibling.ParentID
        if refreshToPtr {
            sibling.Lft = sibling.Lft + 2
            sibling.Rght = sibling.Rght + 2
        }
    case Right:
        sibling := existCtx.ModelBase
        edge = sibling.Rght + 1
        base.Lvl = sibling.Lvl
        base.ParentID = sibling.ParentID
    default:
        return UnsupportedPositionError
    }

    base.Lft = edge
    base.Rght = edge + 1
    spaceTarget := base.Lft - 1

    base.TreeID = existCtx.ModelBase.TreeID

    err := db.createSpace(n, 2, spaceTarget, base.TreeID)
    if err != nil {
        return err
    }
    db.setModelBase(n, base)
    if refreshToPtr {
        db.setModelBase(toPtr, existCtx.ModelBase)
    }
    return db.Statement.Create(n).Error
}
