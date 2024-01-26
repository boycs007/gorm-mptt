package tests

import (
	mptt "github.com/boycs007/gorm-mptt"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	rebuildManager mptt.TreeManager
	rawCreate      = func(manager mptt.TreeManager, node *Node) (int, error) {
		dbNode := &CustomTree{
			ModelBase: mptt.ModelBase{
				ParentID: node.ParentID,
			},
			Name: node.Name,
		}
		err := manager.GormDB().Model(new(CustomTree)).Create(dbNode).Error
		return dbNode.ID, err
	}
	rebuildTestcase = []struct {
		name  string
		level int
		left  int
		right int
	}{
		{
			name:  "dev department",
			level: 1,
			left:  1,
			right: 30,
		},
		{
			name:  "dev center",
			level: 2,
			left:  2,
			right: 15,
		},
		{
			name:  "dev group 1",
			level: 3,
			left:  3,
			right: 8,
		},
		{
			name:  "dev team 1",
			level: 4,
			left:  4,
			right: 5,
		},
		{
			name:  "dev team 2",
			level: 4,
			left:  6,
			right: 7,
		},
	}
)

func caseRebuildBefore(t *testing.T) {
	if rebuildManager != nil {
		return
	}
	var err error
	rebuildManager, err = mptt.NewTreeManager(globalDb, new(CustomTree))
	assert.Nil(t, err)
}

func TestDFSRebuild(t *testing.T) {
	caseRebuildBefore(t)
	for _, node := range rawNodes {
		err := dfs(node, func(n *Node) (int, error) {
			return rawCreate(rebuildManager, n)
		})
		assert.Nil(t, err)
	}
	err := rebuildManager.Rebuild()
	assert.Nil(t, err)
	nodeMap, err := getAllNodes(rebuildManager)
	assert.Nil(t, err)

	for _, testcase := range rebuildTestcase {
		node, ok := nodeMap[testcase.name]
		assert.True(t, ok)
		assert.EqualValues(t, testcase.level, node.Lvl)
		assert.EqualValues(t, testcase.left, node.Lft)
		assert.EqualValues(t, testcase.right, node.Rght)
	}
}

func TestBFSRebuild(t *testing.T) {
	caseRebuildBefore(t)
	for _, node := range rawNodes {
		err := bfs(node, func(n *Node) (int, error) {
			return rawCreate(rebuildManager, n)
		})
		assert.Nil(t, err)
	}
	err := rebuildManager.Rebuild()
	assert.Nil(t, err)
	nodeMap, err := getAllNodes(rebuildManager)
	assert.Nil(t, err)

	for _, testcase := range rebuildTestcase {
		node, ok := nodeMap[testcase.name]
		assert.True(t, ok)
		assert.EqualValues(t, testcase.level, node.Lvl)
		assert.EqualValues(t, testcase.left, node.Lft)
		assert.EqualValues(t, testcase.right, node.Rght)
	}
}
