package mptt

import (
    "reflect"

    "gorm.io/gorm"
)

func (db *Tree) GetTableName(n interface{}) string {
    t := reflect.TypeOf(n)
    if t.Kind() == reflect.Ptr {
        t = t.Elem()
    }
    return db.NamingStrategy.TableName(t.Name())
}

func (db *Tree) validateType(n interface{}) error {
    kind := reflect.TypeOf(n).Kind()
    if kind != reflect.Ptr {
        return ModelTypeError
    }
    return nil
}

func (db *Tree) getContext(nodePtr interface{}) TreeContext {
    mb := db.getModelBase(nodePtr)
    ctx := TreeContext{
        nodePtr,
        &mb,
    }
    return ctx
}

func (db *Tree) getModelBase(n interface{}) MPTTModelBase {
    kind := reflect.TypeOf(n).Kind()
    rv := reflect.ValueOf(n)
    if kind == reflect.Ptr {
        rv = rv.Elem()
    }
    return rv.FieldByName("MPTTModelBase").Interface().(MPTTModelBase)
}

func (db *Tree) setModelBase(n interface{}, base *MPTTModelBase) {
    v := reflect.ValueOf(n).Elem()
    v.FieldByName("MPTTModelBase").Set(reflect.ValueOf(base).Elem())
}

func (db *Tree) getNodeById(ctx TreeContext) *MPTTModelBase {
    base := ctx.ModelBase
    result := &MPTTModelBase{}
    db.Statement.DB.Model(ctx.Node).Where("id = ?", base.ID).First(result)
    return result

}
func (db *Tree) getNodeByParentId(ctx TreeContext) *MPTTModelBase {
    base := ctx.ModelBase
    result := &MPTTModelBase{}
    db.Statement.DB.Model(ctx.Node).Where("id = ?", base.ParentID).First(result)
    return result
}

func (db *Tree) getTreeId(ctx TreeContext) int {
    base := ctx.ModelBase
    if base.ParentID == 0 {
        return db.getNextTreeId(ctx.Node)
    }
    pb := db.getNodeByParentId(ctx)
    return pb.TreeID
}

func (db *Tree) getNextTreeId(nPtr interface{}) int {
    var treeId int
    db.Statement.Select("tree_id").
        Model(nPtr).Order("tree_id desc").Limit(1).Scan(&treeId)
    return treeId + 1
}

func (db *Tree) getMax(ctx TreeContext) int {
    var rght int
    db.Statement.Select("rght").Model(ctx.Node).
        Where("tree_id = ?", ctx.ModelBase.TreeID).Order("rght desc").Limit(1).Scan(&rght)
    return rght
}

func (db *Tree) getLftFromTargetNode(ctx TreeContext, pos int) int {
    var lft int
    db.Statement.DB.Model(ctx.Node).Select("lft").
        Where("parent_id = ?", ctx.ModelBase.ParentID).
        Where("rght < ?", ctx.ModelBase.Lft).
        Order("lft desc").
        Limit(1).
        Offset(pos - 1).Scan(&lft)
    return lft
}

func (db *Tree) getRghtFromTargetNode(ctx TreeContext, pos int) int {
    var rght int
    db.Statement.DB.Model(ctx.Node).Select("rght").
        Where("parent_id = ?", ctx.ModelBase.ParentID).
        Where("lft > ?", ctx.ModelBase.Rght).
        Order("lft asc").
        Limit(1).
        Offset(pos - 1).Scan(&rght)
    return rght
}

func (db *Tree) createTreeSpace(model interface{}, targetTreeId, num int) error {
    return db.Statement.DB.Model(model).
        Where("tree_id > ?", targetTreeId).
        Update("tree_id", gorm.Expr("tree_id + ?", num)).Error
}
