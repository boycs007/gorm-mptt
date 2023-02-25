package mptt

import "gorm.io/gorm"

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

    err = db.Statement.DB.Table(db.GetTableName(n)).
        Where("tree_id = ? AND lft >= ? AND lft < ?", realNode.TreeID, realNode.Lft, realNode.Rght).
        Delete(map[string]interface{}{}).Error
    if err != nil {
        return err
    }
    if realNode.ParentID == 0 {
        // delete the whole tree, close the tree id gap.
        return db.Statement.DB.Table(db.GetTableName(n)).
            Where("tree_id > ?", realNode.TreeID).
            Update("tree_id", gorm.Expr("tree_id - ?", 1)).Error
    }
    return db.closeGap(n, diff, realNode.Rght, realNode.TreeID)
}
