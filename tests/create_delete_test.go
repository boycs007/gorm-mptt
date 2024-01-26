package tests

import (
	"fmt"
	mptt "github.com/boycs007/gorm-mptt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_MutilTrees(t *testing.T) {
	manager, _ := mptt.NewTreeManager(globalDb, new(CustomTree))

	println("create root nodes")
	roots := make([]*CustomTree, 0)
	for i := range make([]struct{}, 10) {
		node := &CustomTree{
			Name: fmt.Sprintf("RootNode%d", i+1),
		}
		err := manager.CreateNode(node)
		assert.Nil(t, err, "Create Node failed: %s", err)
		assert.Equal(t, node.TreeID, i+1)
		assert.Equal(t, node.Lft, 1)
		assert.Equal(t, node.Rght, 2)
		assert.Equal(t, node.Lvl, 1)
		roots = append(roots, node)
	}

	println("create last child")
	depts := make([]*CustomTree, 0)
	for j := range make([]struct{}, 5) {
		subNode := &CustomTree{
			Name: fmt.Sprintf("DeptNode%d", j+1),
		}
		err := manager.InsertNode(subNode, roots[4], mptt.LastChild)
		assert.Nil(t, err, "Insert Node failed: %s", err)
		assert.Equal(t, roots[4].Rght, 2+2*(j+1))
		assert.Equal(t, subNode.TreeID, roots[4].TreeID)
		assert.Equal(t, subNode.Lft, roots[4].Lft+(j*2)+1)

		depts = append(depts, subNode)
	}

	println("create first child")
	secondNode := &CustomTree{
		Name: "SecondGroup",
	}
	err := manager.InsertNode(secondNode, depts[3], mptt.FirstChild)
	assert.Nil(t, err, "Insert Node failed: %s", err)

	firstNode := &CustomTree{
		Name: "FirstGroup",
	}
	err = manager.InsertNode(firstNode, depts[3], mptt.FirstChild)
	assert.Nil(t, err, "Insert Node failed: %s", err)

	err = manager.RefreshNode(secondNode)
	assert.Nil(t, err)
	assert.Equal(t, depts[3].Lft+3, secondNode.Lft)
	assert.Equal(t, depts[3].Lft+4, secondNode.Rght)
	assert.Equal(t, depts[3].Lvl+1, secondNode.Lvl)

	firstDept := &CustomTree{
		Name: "FirstDept",
	}
	err = manager.InsertNode(firstDept, roots[4], mptt.FirstChild)
	assert.Nil(t, err, "Insert Node failed: %s", err)

	_ = manager.RefreshNode(roots[2])

	err = manager.InsertNode(&CustomTree{
		Name: "InsertLeftRoot",
	}, roots[2], mptt.Left)
	assert.Nil(t, err, "Insert Node failed: %s", err)

	_ = manager.RefreshNode(depts[2])

	err = manager.InsertNode(&CustomTree{
		Name: "InsertLeftDept",
	}, depts[2], mptt.Left)
	assert.Nil(t, err, "Insert Node failed: %s", err)

	err = manager.DeleteNode(roots[5])
	assert.Nil(t, err, "Delete Node failed: %s", err)
	err = manager.RefreshNode(roots[7])
	assert.Nil(t, err, "GetNode Node failed: %s", err)
	assert.Equal(t, 8, roots[7].TreeID)
	outPtr := &CustomTree{}
	err = manager.Node(roots[7]).GetRoot(outPtr)
	assert.Nil(t, err, "GetRoot Node failed: %s", err)
	assert.Equal(t, 8, outPtr.ID)

	err = manager.DeleteNode(depts[1])
	assert.Nil(t, err, "GetRoot Node failed: %s", err)
	refreshDb()
}

func TestDFSCreate(t *testing.T) {
	manager, err := mptt.NewTreeManager(globalDb, new(CustomTree))
	assert.Nil(t, err)
	var f = func(node *Node) (int, error) {
		return createNode(manager, node)
	}
	for _, node := range rawNodes {
		err = dfs(node, f)
		assert.Nil(t, err)
	}
	// verify data by view database

	item, err := getItemByName(manager, "dev team 4")
	assert.Nil(t, err)
	assert.EqualValues(t, item.Lvl, 4)
	assert.EqualValues(t, item.Lft, 12)
	assert.EqualValues(t, item.Rght, 13)

	item, err = getItemByName(manager, "design group 2")
	assert.Nil(t, err)
	assert.EqualValues(t, item.Lvl, 3)
	assert.EqualValues(t, item.Lft, 23)
	assert.EqualValues(t, item.Rght, 28)
}

func TestBFSCreate(t *testing.T) {
	manager, err := mptt.NewTreeManager(globalDb, new(CustomTree))
	assert.Nil(t, err)
	var f = func(node *Node) (int, error) {
		return createNode(manager, node)
	}
	for _, node := range rawNodes {
		err = bfs(node, f)
		assert.Nil(t, err)
	}
	// verify data by view database
	item, err := getItemByName(manager, "dev team 4")
	assert.Nil(t, err)
	assert.EqualValues(t, item.Lvl, 4)
	assert.EqualValues(t, item.Lft, 12)
	assert.EqualValues(t, item.Rght, 13)

	item, err = getItemByName(manager, "design group 2")
	assert.Nil(t, err)
	assert.EqualValues(t, item.Lvl, 3)
	assert.EqualValues(t, item.Lft, 23)
	assert.EqualValues(t, item.Rght, 28)
}
