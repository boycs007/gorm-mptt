package tests

import (
	mptt "github.com/boycs007/gorm-mptt"
	"gorm.io/gorm/clause"
)

func dfs(node *Node, f func(n *Node) (int, error)) error {
	if node == nil {
		return nil
	}
	id, err := f(node)
	if err != nil {
		return err
	}
	for _, child := range node.Children {
		child.ParentID = id
		if err = dfs(child, f); err != nil {
			return err
		}
	}
	return nil
}

func bfs(root *Node, f func(n *Node) (int, error)) error {
	if root == nil {
		return nil
	}

	queue := []*Node{root}

	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		id, err := f(node)
		if err != nil {
			return err
		}
		for _, child := range node.Children {
			child.ParentID = id
			queue = append(queue, child)
		}
	}
	return nil
}

func createNode(manager mptt.TreeManager, node *Node) (int, error) {
	ctNode := &CustomTree{
		ModelBase: mptt.ModelBase{
			ParentID: node.ParentID,
		},
		Name: node.Name,
	}
	e := manager.CreateNode(ctNode)
	return ctNode.ID, e
}

func getItemByName(manager mptt.TreeManager, name string) (*CustomTree, error) {
	var ret CustomTree
	err := manager.GormDB().Model(&ret).Where(clause.Eq{
		Column: "name",
		Value:  name,
	}).First(&ret).Error
	return &ret, err
}

func getAllNodes(manager mptt.TreeManager) (map[string]*CustomTree, error) {
	var ret []*CustomTree
	err := manager.GormDB().Model(new(CustomTree)).Find(&ret).Error
	retMap := make(map[string]*CustomTree)
	for _, item := range ret {
		retMap[item.Name] = item
	}
	return retMap, err
}
