package mptt

import "gorm.io/gorm/clause"

func (t *tree) GetJoinClause(tableName string, colName string) clause.Join {
	return clause.Join{
		Type:  clause.LeftJoin,
		Table: clause.Table{Name: t.tableName},
		ON: clause.Where{
			Exprs: []clause.Expression{clause.Eq{Column: clause.Column{Table: tableName, Name: colName}, Value: clause.PrimaryColumn}},
		},
	}
}

func (t *tree) GetAncestorsClause(rawItem interface{}, includeSelf bool) clause.Where {
	cls := clause.Where{
		Exprs: []clause.Expression{
			clause.Eq{Column: t.colTree(true), Value: t.getTreeID(rawItem)},
		},
	}
	if includeSelf {
		cls.Exprs = append(cls.Exprs,
			clause.Lte{Column: t.colLeft(true), Value: t.getLeft(rawItem)},
			clause.Gte{Column: t.colRight(true), Value: t.getRight(rawItem)},
		)
	} else {
		cls.Exprs = append(cls.Exprs,
			clause.Lt{Column: t.colLeft(true), Value: t.getLeft(rawItem)},
			clause.Gt{Column: t.colRight(true), Value: t.getRight(rawItem)},
		)
	}
	return cls
}

func (t *tree) GetDescendantsClause(rawItem interface{}, includeSelf bool) clause.Where {
	cls := clause.Where{
		Exprs: []clause.Expression{
			clause.Eq{Column: t.colTree(true), Value: t.getTreeID(rawItem)},
		},
	}
	if includeSelf {
		cls.Exprs = append(cls.Exprs,
			clause.Gte{Column: t.colLeft(true), Value: t.getLeft(rawItem)},
			clause.Lte{Column: t.colRight(true), Value: t.getRight(rawItem)},
		)
	} else {
		cls.Exprs = append(cls.Exprs,
			clause.Gt{Column: t.colLeft(true), Value: t.getLeft(rawItem)},
			clause.Lt{Column: t.colRight(true), Value: t.getRight(rawItem)},
		)
	}
	return cls
}
