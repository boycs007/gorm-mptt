package tests

import (
	mptt "github.com/boycs007/gorm-mptt"
	"github.com/stretchr/testify/assert"
	"testing"
)

/**
 * @Description  测试查询节点信息
 * @Author BrookeChen
 * @Date 2024/2/3 21:01
 **/

var (
	queryManager  mptt.TreeManager
	allNodeByName map[string]*CustomTree
)

func caseBefore(t *testing.T) {
	if queryManager != nil {
		return
	}
	var err error
	queryManager, err = mptt.NewTreeManager(globalDb, new(CustomTree))
	assert.Nil(t, err)
	var f = func(node *Node) (int, error) {
		return createNode(queryManager, node)
	}
	for _, node := range rawNodes {
		err = dfs(node, f)
		assert.Nil(t, err)
	}
	allNodeByName, err = getAllNodes(queryManager)
	assert.Nil(t, err)
}

func Test_IsRootNode(t *testing.T) {
	caseBefore(t)

	testcases := []struct {
		node *CustomTree
		want bool
	}{
		{
			node: allNodeByName["dev department"],
			want: true,
		},
		{
			node: allNodeByName["dev center"],
			want: false,
		}, {
			node: allNodeByName["dev group 1"],
			want: false,
		}, {
			node: allNodeByName["dev team 1"],
			want: false,
		},
	}

	for _, testcase := range testcases {
		ok := queryManager.Node(testcase.node).IsRootNode()
		assert.EqualValues(t, testcase.want, ok)
	}
}

func Test_IsChildNode(t *testing.T) {
	caseBefore(t)

	testcases := []struct {
		node *CustomTree
		want bool
	}{
		{
			node: allNodeByName["dev department"],
			want: false,
		},
		{
			node: allNodeByName["dev center"],
			want: true,
		}, {
			node: allNodeByName["dev group 1"],
			want: true,
		}, {
			node: allNodeByName["dev team 1"],
			want: true,
		},
	}

	for _, testcase := range testcases {
		ok := queryManager.Node(testcase.node).IsChildNode()
		assert.EqualValues(t, testcase.want, ok)
	}
}

func Test_IsLeafNode(t *testing.T) {
	caseBefore(t)

	testcases := []struct {
		node *CustomTree
		want bool
	}{
		{
			node: allNodeByName["dev department"],
			want: false,
		},
		{
			node: allNodeByName["dev center"],
			want: false,
		}, {
			node: allNodeByName["dev group 1"],
			want: false,
		}, {
			node: allNodeByName["dev team 1"],
			want: true,
		},
	}

	for _, testcase := range testcases {
		ok := queryManager.Node(testcase.node).IsLeafNode()
		assert.EqualValues(t, testcase.want, ok)
	}
}

func Test_GetLevel(t *testing.T) {
	caseBefore(t)

	testcases := []struct {
		node *CustomTree
		want int
	}{
		{
			node: allNodeByName["dev department"],
			want: 1,
		},
		{
			node: allNodeByName["dev center"],
			want: 2,
		}, {
			node: allNodeByName["dev group 1"],
			want: 3,
		}, {
			node: allNodeByName["dev team 1"],
			want: 4,
		},
	}

	for _, testcase := range testcases {
		lvl := queryManager.Node(testcase.node).GetLevel()
		assert.EqualValues(t, testcase.want, lvl)
	}
}

func Test_GetDescendantCount(t *testing.T) {
	caseBefore(t)

	testcases := []struct {
		node *CustomTree
		want int
	}{
		{
			node: allNodeByName["dev department"],
			want: 14,
		},
		{
			node: allNodeByName["dev center"],
			want: 6,
		}, {
			node: allNodeByName["dev group 1"],
			want: 2,
		}, {
			node: allNodeByName["dev team 1"],
			want: 0,
		},
	}

	for _, testcase := range testcases {
		count := queryManager.Node(testcase.node).GetDescendantCount()
		assert.EqualValues(t, testcase.want, count)
	}
}

func Test_IsAncestorOf(t *testing.T) {
	caseBefore(t)

	testcases := []struct {
		node            *CustomTree
		targetNode      *CustomTree
		want            bool
		includeSelfWant bool
	}{
		{
			node:            allNodeByName["dev department"],
			targetNode:      allNodeByName["dev department"],
			want:            false,
			includeSelfWant: true,
		},
		{
			node:            allNodeByName["dev center"],
			targetNode:      allNodeByName["dev team 1"],
			want:            true,
			includeSelfWant: true,
		},
		{
			node:            allNodeByName["dev center"],
			targetNode:      allNodeByName["dev department"],
			want:            false,
			includeSelfWant: false,
		},
		{
			node:            allNodeByName["dev department"],
			targetNode:      allNodeByName["dev group 1"],
			want:            true,
			includeSelfWant: true,
		},
	}

	for _, testcase := range testcases {
		ok := queryManager.Node(testcase.node).IsAncestorOf(testcase.targetNode, false)
		assert.EqualValues(t, testcase.want, ok)

		ok = queryManager.Node(testcase.node).IsAncestorOf(testcase.targetNode, true)
		assert.EqualValues(t, testcase.includeSelfWant, ok)

		ok = queryManager.Node(testcase.targetNode).IsDescendantOf(testcase.node, false)
		assert.EqualValues(t, testcase.want, ok)

		ok = queryManager.Node(testcase.targetNode).IsDescendantOf(testcase.node, true)
		assert.EqualValues(t, testcase.includeSelfWant, ok)
	}
}

func Test_GetRoot(t *testing.T) {
	caseBefore(t)

	testRoot := allNodeByName["dev department"]
	testcases := []struct {
		node   *CustomTree
		rootID int
	}{
		{
			node:   allNodeByName["dev department"],
			rootID: testRoot.ID,
		},
		{
			node:   allNodeByName["dev center"],
			rootID: testRoot.ID,
		},
		{
			node:   allNodeByName["dev group 1"],
			rootID: testRoot.ID,
		},
		{
			node:   allNodeByName["dev team 1"],
			rootID: testRoot.ID,
		},
	}

	for _, testcase := range testcases {
		var node CustomTree
		err := queryManager.Node(testcase.node).GetRoot(&node)
		assert.Nil(t, err)
		assert.EqualValues(t, testcase.rootID, node.ID)
	}
}

func Test_GetLeafNodes(t *testing.T) {
	caseBefore(t)
	testcases := []struct {
		node   *CustomTree
		leaves map[string]struct{}
	}{
		{
			node: allNodeByName["dev department"],
			leaves: map[string]struct{}{
				"dev team 1":  {},
				"dev team 2":  {},
				"dev team 3":  {},
				"dev team 4":  {},
				"test team 1": {},
				"test team 2": {},
				"test team 3": {},
				"test team 4": {},
			},
		},
		{
			node: allNodeByName["dev center"],
			leaves: map[string]struct{}{
				"dev team 1": {},
				"dev team 2": {},
				"dev team 3": {},
				"dev team 4": {},
			},
		},
		{
			node: allNodeByName["dev group 1"],
			leaves: map[string]struct{}{
				"dev team 1": {},
				"dev team 2": {},
			},
		},
		{
			node:   allNodeByName["dev team 1"],
			leaves: map[string]struct{}{},
		},
	}

	for _, testcase := range testcases {
		var nodes []*CustomTree
		err := queryManager.Node(testcase.node).GetLeafNodes(&nodes)
		assert.Nil(t, err)
		assert.EqualValues(t, len(testcase.leaves), len(nodes))
		for _, node := range nodes {
			_, ok := testcase.leaves[node.Name]
			assert.True(t, ok)
		}
	}
}

func Test_GetChildren(t *testing.T) {
	caseBefore(t)
	testcases := []struct {
		node *CustomTree
		want map[string]struct{}
	}{
		{
			node: allNodeByName["dev department"],
			want: map[string]struct{}{
				"dev center":  {},
				"test center": {},
			},
		},
		{
			node: allNodeByName["dev center"],
			want: map[string]struct{}{
				"dev group 1": {},
				"dev group 2": {},
			},
		},
		{
			node: allNodeByName["dev group 1"],
			want: map[string]struct{}{
				"dev team 1": {},
				"dev team 2": {},
			},
		},
		{
			node: allNodeByName["dev team 1"],
			want: map[string]struct{}{},
		},
	}

	for _, testcase := range testcases {
		var nodes []*CustomTree
		err := queryManager.Node(testcase.node).GetChildren(&nodes)
		assert.Nil(t, err)
		assert.EqualValues(t, len(testcase.want), len(nodes))
		for _, node := range nodes {
			_, ok := testcase.want[node.Name]
			assert.True(t, ok)
		}
	}
}

func Test_GetSiblings(t *testing.T) {
	caseBefore(t)
	testcases := []struct {
		node *CustomTree
		want map[string]struct{}
	}{
		{
			node: allNodeByName["dev department"],
			want: map[string]struct{}{
				"product department": {},
			},
		},
		{
			node: allNodeByName["dev center"],
			want: map[string]struct{}{
				"test center": {},
			},
		},
		{
			node: allNodeByName["dev group 1"],
			want: map[string]struct{}{
				"dev group 2": {},
			},
		},
		{
			node: allNodeByName["dev team 1"],
			want: map[string]struct{}{
				"dev team 2": {},
			},
		},
	}

	for _, testcase := range testcases {
		var nodes []*CustomTree
		err := queryManager.Node(testcase.node).GetSiblings(&nodes, false)
		assert.Nil(t, err)
		assert.EqualValues(t, len(testcase.want), len(nodes))
		for _, node := range nodes {
			_, ok := testcase.want[node.Name]
			assert.True(t, ok)
		}
		err = queryManager.Node(testcase.node).GetSiblings(&nodes, true)
		assert.Nil(t, err)
		assert.EqualValues(t, len(testcase.want)+1, len(nodes))
		testcase.want[testcase.node.Name] = struct{}{}
		for _, node := range nodes {
			_, ok := testcase.want[node.Name]
			assert.True(t, ok)
		}
	}
}

func Test_GetDescendants(t *testing.T) {
	caseBefore(t)
	testcases := []struct {
		node *CustomTree
		want map[string]struct{}
	}{
		{
			node: allNodeByName["dev department"],
			want: map[string]struct{}{
				"dev center":   {},
				"dev group 1":  {},
				"dev team 1":   {},
				"dev team 2":   {},
				"dev group 2":  {},
				"dev team 3":   {},
				"dev team 4":   {},
				"test center":  {},
				"test group 1": {},
				"test team 1":  {},
				"test team 2":  {},
				"test group 2": {},
				"test team 3":  {},
				"test team 4":  {},
			},
		},
		{
			node: allNodeByName["dev center"],
			want: map[string]struct{}{
				"dev group 1": {},
				"dev team 1":  {},
				"dev team 2":  {},
				"dev group 2": {},
				"dev team 3":  {},
				"dev team 4":  {},
			},
		},
		{
			node: allNodeByName["dev group 1"],
			want: map[string]struct{}{
				"dev team 1": {},
				"dev team 2": {},
			},
		},
		{
			node: allNodeByName["dev team 1"],
			want: map[string]struct{}{},
		},
	}

	for _, testcase := range testcases {
		var nodes []*CustomTree
		err := queryManager.Node(testcase.node).GetDescendants(&nodes, false)
		assert.Nil(t, err)
		assert.EqualValues(t, len(testcase.want), len(nodes))
		for _, node := range nodes {
			_, ok := testcase.want[node.Name]
			assert.True(t, ok)
		}
		err = queryManager.Node(testcase.node).GetDescendants(&nodes, true)
		assert.Nil(t, err)
		assert.EqualValues(t, len(testcase.want)+1, len(nodes))
		testcase.want[testcase.node.Name] = struct{}{}
		for _, node := range nodes {
			_, ok := testcase.want[node.Name]
			assert.True(t, ok)
		}
	}
}

func Test_GetAncestors(t *testing.T) {
	caseBefore(t)
	testcases := []struct {
		node *CustomTree
		want []string
	}{
		{
			node: allNodeByName["dev department"],
			want: []string{},
		},
		{
			node: allNodeByName["dev center"],
			want: []string{
				"dev department",
			},
		},
		{
			node: allNodeByName["dev group 1"],
			want: []string{
				"dev department",
				"dev center",
			},
		},
		{
			node: allNodeByName["dev team 1"],
			want: []string{
				"dev department",
				"dev center",
				"dev group 1",
			},
		},
	}

	for _, testcase := range testcases {
		var nodes []*CustomTree
		err := queryManager.Node(testcase.node).GetAncestors(&nodes, true, false)
		assert.Nil(t, err)
		assert.EqualValues(t, len(testcase.want), len(nodes))
		for idx, node := range nodes {
			assert.EqualValues(t, testcase.want[idx], node.Name)
		}
		err = queryManager.Node(testcase.node).GetAncestors(&nodes, true, true)
		assert.Nil(t, err)
		assert.EqualValues(t, len(testcase.want)+1, len(nodes))
		wants := append(testcase.want, testcase.node.Name)
		for idx, node := range nodes {
			assert.EqualValues(t, wants[idx], node.Name)
		}
	}
}

func Test_GetFamily(t *testing.T) {
	caseBefore(t)
	testcases := []struct {
		node *CustomTree
		want map[string]struct{}
	}{
		{
			node: allNodeByName["dev department"],
			want: map[string]struct{}{
				"dev department": {},
				"dev center":     {},
				"dev group 1":    {},
				"dev team 1":     {},
				"dev team 2":     {},
				"dev group 2":    {},
				"dev team 3":     {},
				"dev team 4":     {},
				"test center":    {},
				"test group 1":   {},
				"test team 1":    {},
				"test team 2":    {},
				"test group 2":   {},
				"test team 3":    {},
				"test team 4":    {},
			},
		},
		{
			node: allNodeByName["dev center"],
			want: map[string]struct{}{
				"dev department": {},
				"dev center":     {},
				"dev group 1":    {},
				"dev team 1":     {},
				"dev team 2":     {},
				"dev group 2":    {},
				"dev team 3":     {},
				"dev team 4":     {},
			},
		},
		{
			node: allNodeByName["dev group 1"],
			want: map[string]struct{}{
				"dev department": {},
				"dev center":     {},
				"dev group 1":    {},
				"dev team 1":     {},
				"dev team 2":     {},
			},
		},
		{
			node: allNodeByName["dev team 1"],
			want: map[string]struct{}{
				"dev department": {},
				"dev center":     {},
				"dev group 1":    {},
				"dev team 1":     {},
			},
		},
	}

	for _, testcase := range testcases {
		var nodes []*CustomTree
		err := queryManager.Node(testcase.node).GetFamily(&nodes)
		assert.Nil(t, err)
		assert.EqualValues(t, len(testcase.want), len(nodes))
		for _, node := range nodes {
			_, ok := testcase.want[node.Name]
			assert.True(t, ok)
		}
	}
}

func Test_GetNextSibling(t *testing.T) {
	caseBefore(t)
	testcases := []struct {
		node *CustomTree
		want string
	}{
		{
			node: allNodeByName["dev department"],
			want: "product department",
		},
		{
			node: allNodeByName["dev center"],
			want: "test center",
		},
		{
			node: allNodeByName["dev group 1"],
			want: "dev group 2",
		},
		{
			node: allNodeByName["dev team 1"],
			want: "dev team 2",
		},
	}

	for _, testcase := range testcases {
		var node CustomTree
		err := queryManager.Node(testcase.node).GetNextSibling(&node)
		assert.Nil(t, err)
		assert.EqualValues(t, testcase.want, node.Name)
	}
}

func Test_GetPreviousSibling(t *testing.T) {
	caseBefore(t)
	testcases := []struct {
		node *CustomTree
		want string
	}{
		{
			node: allNodeByName["product department"],
			want: "dev department",
		},
		{
			node: allNodeByName["test center"],
			want: "dev center",
		},
		{
			node: allNodeByName["dev group 2"],
			want: "dev group 1",
		},
		{
			node: allNodeByName["dev team 2"],
			want: "dev team 1",
		},
	}

	for _, testcase := range testcases {
		var node CustomTree
		err := queryManager.Node(testcase.node).GetPreviousSibling(&node)
		assert.Nil(t, err)
		assert.EqualValues(t, testcase.want, node.Name)
	}
}
