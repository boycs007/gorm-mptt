package mptt

// DeleteNode delete current node and all descendants
func (db *Tree) DeleteNode(n interface{}) error {
    var err error
    if err = db.validateType(n); err != nil {
        return err
    }
    ctx := db.getContext(n)
    realNode := db.getNodeById(ctx)
    if realNode.ID == 0 {
        return NodeNotExistsError
    }
    diff := realNode.Rght - realNode.Lft + 1

    result := map[string]interface{}{}
    err = db.Statement.DB.Model(db.Context.Node).
        Where("tree_id = ? AND lft >= ? AND lft < ?",
            realNode.TreeID, realNode.Lft, realNode.Rght).Delete(&result).Error
    if err != nil {
        return err
    }
    return db.closeGap(n, diff, realNode.Rght, realNode.TreeID)
}
