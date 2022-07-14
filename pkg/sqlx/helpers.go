package sqlx

import (
	"github.com/octohelm/storage/pkg/sqlbuilder"
)

func InsertToDB(db DBExecutor, model sqlbuilder.Model, zeroFields []string, additions ...sqlbuilder.Addition) sqlbuilder.SqlExpr {
	table := db.T(model)
	cols, vals := sqlbuilder.ColumnsAndValuesByFieldValues(table, sqlbuilder.FieldValuesFromStructByNonZero(model, zeroFields...))
	return sqlbuilder.Insert().Into(table, additions...).Values(cols, vals...)
}

func AsAssignments(db DBExecutor, model sqlbuilder.Model, zeroFields ...string) sqlbuilder.Assignments {
	table := db.T(model)
	return sqlbuilder.AssignmentsByFieldValues(table, sqlbuilder.FieldValuesFromStructByNonZero(model, zeroFields...))
}
