package mptt

// CreateNode 插入新节点。当节点的ParentID为0时，将生成新的树的根节点；
// 当节点的ParentID不为0时，将新节点插入为Parent的最后一个子节点
func (t *tree) CreateNode(n interface{}) error {
	var err error
	if err = t.validateType(n); err != nil {
		return err
	}
	parentID := t.getParentID(n)
	if isEmpty(parentID) {
		// new tree root node
		t.setTreeID(n, t.getNextTreeId())
		t.setLeft(n, 1)
		t.setRight(n, 2)
		t.setLevel(n, 1)
		return t.Statement.Create(n).Error
	}
	parent, err := t.getNodeByID(parentID)
	if err != nil {
		return err
	}
	return t.InsertNode(n, parent, LastChild)
}

// InsertNode 插入新节点
//
// @param n: 需要插入的节点
// @param toPtr: 需要插入到哪里，toPtr为已存在的某个节点
// @param position:  需要插入到相对于toPtr的某个位置，可以是其
//
//	LastChild: 作为toPtr的最后一个子节点
//	FirstChild: 作为toPtr的第一个子节点
//	Left: 插入到toPtr的左边(前面)
//	Right: 插入到toPtr的右边(后面)
//
// @param refreshToPtr: 是否需要将toPtr对象的信息进行更新，例如如果插入到toPtr的左侧后，toPtr的lft、rght值将会更新
func (t *tree) InsertNode(n, toPtr interface{}, position PositionEnum) error {
	var (
		err  error
		edge int
	)
	if err = t.validateType(n); err != nil {
		return err
	}
	if err = t.validateType(toPtr); err != nil {
		return err
	}
	var (
		existLvl      = t.getLevel(toPtr)
		existLeft     = t.getLeft(toPtr)
		existRight    = t.getRight(toPtr)
		existID       = t.getNodeID(toPtr)
		existParentID = t.getParentID(toPtr)
		existTreeID   = t.getTreeID(toPtr)
	)
	// 根节点插入
	if isEmpty(existParentID) && (position == Left || position == Right) {
		var spaceTarget int
		if position == Left {
			t.setTreeID(n, existTreeID)
			spaceTarget = existTreeID - 1
			t.setTreeID(toPtr, existTreeID+1)
		} else {
			t.setTreeID(n, existTreeID+1)
			spaceTarget = existTreeID
		}
		t.setLeft(n, 1)
		t.setRight(n, 2)
		t.setLevel(n, 1)
		if err = t.createTreeSpace(n, spaceTarget, 1); err != nil {
			return err
		}
		return t.Statement.Create(n).Error
	}

	switch position {
	case LastChild:
		// toPtr is node's parent
		edge = existRight
		t.setLevel(n, existLvl+1)
		t.setParentID(n, existID)
		// refresh father
		t.setRight(toPtr, existRight+2)
	case FirstChild:
		// toPtr is node's parent
		edge = existLeft + 1
		t.setLevel(n, existLvl+1)
		t.setParentID(n, existID)
		// refresh father
		t.setRight(toPtr, existRight+2)
	case Left:
		// toPtr is node's first sibling
		edge = existLeft
		t.setLevel(n, existLvl)
		t.setParentID(n, existParentID)
		// refresh exist node
		t.setLeft(toPtr, existLeft+2)
		t.setRight(toPtr, existRight+2)
	case Right:
		edge = existRight + 1
		t.setLevel(n, existLvl)
		t.setParentID(n, existParentID)
	default:
		return UnsupportedPositionError
	}

	t.setLeft(n, edge)
	t.setRight(n, edge+1)
	spaceTarget := edge - 1

	t.setTreeID(n, existTreeID)

	err = t.createSpace(2, spaceTarget, existTreeID)
	if err != nil {
		return err
	}
	return t.Statement.Create(n).Error
}
