package sqlbuilder

import (
	"context"
)

type OnConflictAddition interface {
	Addition
	DoNothing() OnConflictAddition
	DoUpdateSet(assignments ...Assignment) OnConflictAddition
}

func OnConflict(columns ColumnCollection) OnConflictAddition {
	return &onConflict{
		columns: columns,
	}
}

type onConflict struct {
	columns     ColumnCollection
	doNothing   bool
	assignments []Assignment
}

func (onConflict) AdditionType() AdditionType {
	return AdditionOnConflict
}

func (o onConflict) DoNothing() OnConflictAddition {
	o.doNothing = true
	return &o
}

func (o onConflict) DoUpdateSet(assignments ...Assignment) OnConflictAddition {
	o.assignments = assignments
	return &o
}

func (o *onConflict) IsNil() bool {
	return o == nil || IsNilExpr(o.columns) || (!o.doNothing && len(o.assignments) == 0)
}

func (o *onConflict) Ex(ctx context.Context) *Ex {
	e := Expr("ON CONFLICT ")

	e.WriteGroup(func(e *Ex) {
		e.WriteExpr(o.columns)
	})

	e.WriteQuery(" DO ")

	if o.doNothing {
		e.WriteQuery("NOTHING")
	} else {
		e.WriteQuery("UPDATE SET ")
		for i := range o.assignments {
			if i > 0 {
				e.WriteQuery(", ")
			}
			e.WriteExpr(o.assignments[i])
		}
	}

	return e.Ex(ctx)
}
