package tests

import (
	mptt "github.com/boycs007/gorm-mptt"
	"github.com/stretchr/testify/assert"
	"testing"
)

/**
 * @Description  测试移动节点
 * @Author BrookeChen
 * @Date 2024/2/01 21:01
 **/
func Test_Move(t *testing.T) {
	manager, err := mptt.NewTreeManager(globalDb, new(CustomTree))
	assert.Nil(t, err)
	var f = func(node *Node) (int, error) {
		return createNode(manager, node)
	}
	for _, node := range rawNodes {
		err = dfs(node, f)
		assert.Nil(t, err)
	}
	nodeByName, err := getAllNodes(manager)
	assert.Nil(t, err)
	// move sibling to raw
	node1 := nodeByName["dev team 1"]
	node := nodeByName["dev team 2"]
	ok, err := manager.MoveNode(node, node1, mptt.Right)
	assert.Nil(t, err)
	assert.True(t, ok)
	assert.EqualValues(t, 4, node.Lvl)
	assert.EqualValues(t, 6, node.Lft)
	assert.EqualValues(t, 7, node.Rght)
	// move sibling to last
	ok, err = manager.MoveNode(node1, node, mptt.Right)
	assert.Nil(t, err)
	assert.True(t, ok)
	assert.EqualValues(t, 4, node1.Lvl)
	assert.EqualValues(t, 6, node1.Lft)
	assert.EqualValues(t, 7, node1.Rght)
	// move sibling to first
	ok, err = manager.MoveNode(node1, node, mptt.Left)
	assert.Nil(t, err)
	assert.True(t, ok)
	assert.EqualValues(t, 4, node.Lvl)
	assert.EqualValues(t, 6, node.Lft)
	assert.EqualValues(t, 7, node.Rght)

	// move node to other parent
	// make dev team 2 as the child of test team 1
	node2 := nodeByName["test team 1"]
	ok, err = manager.MoveNode(node1, node2, mptt.FirstChild, true)
	assert.Nil(t, err)
	assert.True(t, ok)
	assert.EqualValues(t, 4, node2.Lvl)
	assert.EqualValues(t, 16, node2.Lft)
	assert.EqualValues(t, 19, node2.Rght)
	assert.EqualValues(t, 5, node1.Lvl)
	assert.EqualValues(t, 17, node1.Lft)
	assert.EqualValues(t, 18, node1.Rght)

	// move sub node to root
	node3 := nodeByName["product department"]
	ok, err = manager.MoveNode(node1, node3, mptt.Right)
	assert.Nil(t, err)
	assert.True(t, ok)
	assert.EqualValues(t, 0, node1.ParentID)
	assert.EqualValues(t, 3, node1.TreeID)
	assert.EqualValues(t, 1, node1.Lvl)
	assert.EqualValues(t, 1, node1.Lft)
	assert.EqualValues(t, 2, node1.Rght)

	// move root node to left
	ok, err = manager.MoveNode(node1, node3, mptt.Left, true)
	assert.EqualValues(t, 0, node1.ParentID)
	assert.EqualValues(t, 2, node1.TreeID)
	assert.EqualValues(t, 3, node3.TreeID)

	// move single root node to child
	node4 := nodeByName["dev group 1"]
	ok, err = manager.MoveNode(node1, node4, mptt.FirstChild, true)
	assert.EqualValues(t, 1, node1.TreeID)
	assert.EqualValues(t, 4, node1.Lvl)
	assert.EqualValues(t, 4, node1.Lft)
	assert.EqualValues(t, 5, node1.Rght)

	// move sub tree to root node
	node5 := nodeByName["dev center"]
	ok, err = manager.MoveNodeByID(node5.ID, node3.ID, mptt.Left)
	assert.Nil(t, err)
	assert.True(t, ok)
	node6, err := getItemByName(manager, "dev team 2")
	assert.Nil(t, err)
	assert.True(t, ok)
	assert.EqualValues(t, 3, node6.TreeID)
	assert.EqualValues(t, 3, node6.Lvl)
	assert.EqualValues(t, 5, node6.Lft)
	assert.EqualValues(t, 6, node6.Rght)
	node7, err := getItemByName(manager, "dev department")
	assert.Nil(t, err)

	// move root tree to child node
	ok, err = manager.MoveNodeByID(node5.ID, node7.ID, mptt.LastChild)
	assert.Nil(t, err)
	assert.True(t, ok)
	node6, err = getItemByName(manager, "dev team 2")
	assert.Nil(t, err)
	assert.True(t, ok)
	// 1,4,20,21
	assert.EqualValues(t, 1, node6.TreeID)
	assert.EqualValues(t, 4, node6.Lvl)
	assert.EqualValues(t, 20, node6.Lft)
	assert.EqualValues(t, 21, node6.Rght)

	// move child tree to child tree
	ok, err = manager.MoveNodeByID(node5.ID, node3.ID, mptt.FirstChild)
	assert.Nil(t, err)
	assert.True(t, ok)

}
