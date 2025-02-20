package extractor

import (
	"context"
	"fmt"
	"iter"
	"slices"
	"strings"

	"github.com/octohelm/storage/deprecated/pkg/dal"
	"github.com/octohelm/storage/pkg/er"
	"github.com/octohelm/storage/pkg/session"
	"github.com/octohelm/storage/pkg/sqlbuilder"
	"github.com/octohelm/storage/pkg/sqlfrag"
	"github.com/octohelm/storage/pkg/sqltype"
)

func FromCatalog(ctx context.Context, s session.Session, tables *sqlbuilder.Tables) *er.OrderedDatabase {
	erd := &er.OrderedDatabase{
		Head: er.Head{
			Name: s.Name(),
		},
	}

	c := &collector{
		s:       s,
		uniques: make(map[string]string),
	}

	for t := range c.tables(ctx, tables) {
		erd.Tables.Set(t.Name, t)
	}

	for _, t := range erd.Tables.KeyValues() {
		for _, col := range t.Columns.KeyValues() {
			if of, ok := c.uniques[col.GoType]; ok {
				if !strings.HasPrefix(of, t.Name+".") {
					col.Of = of
				}
			}
		}
	}

	return erd
}

type collector struct {
	s session.Session

	uniques map[string]string
}

func (c *collector) tables(ctx context.Context, tables *sqlbuilder.Tables) iter.Seq[*er.OrderedTable] {
	return func(yield func(*er.OrderedTable) bool) {
		for table := range tables.Range {
			t := &er.OrderedTable{
				Head: er.Head{
					Name: table.TableName(),
				},
			}

			v, ok := table.(interface{ New() sqlbuilder.Model })
			if !ok {
				continue
			}
			m := v.New()

			for col := range c.columns(ctx, table, m) {
				t.Columns.Set(col.Name, col)
			}

			for cc := range c.constraints(ctx, table, m) {
				t.Constraints.Set(cc.Name, cc)
			}

			c.mayCollectRuntimeDoc(m, &t.Head)

			if !yield(t) {
				return
			}
		}
	}
}

func (c *collector) columns(ctx context.Context, table sqlbuilder.Table, m sqlbuilder.Model) iter.Seq[*er.OrderedColumn] {
	return func(yield func(*er.OrderedColumn) bool) {
		for col := range table.Cols() {
			def := sqlbuilder.GetColumnDef(col)
			if def.DeprecatedActions != nil {
				continue
			}

			c2 := &er.OrderedColumn{
				Head: er.Head{
					Name: col.Name(),
				},
				Type: def.DataType,
			}

			if goType := def.Type.String(); strings.Contains(goType, ".") {
				c2.GoType = goType
			}

			c2.Type, _ = sqlfrag.Collect(ctx, c.s.Adapter().Dialect().DataType(def))

			c.mayCollectRuntimeDoc(m, &c2.Head, col.FieldName())

			if !yield(c2) {
				return
			}
		}
	}
}

func (c *collector) constraints(ctx context.Context, table sqlbuilder.Table, m sqlbuilder.Model) iter.Seq[*er.OrderedConstraint] {
	softDeletedField := ""
	if x, ok := m.(dal.ModelWithSoftDelete); ok {
		softDeletedField, _ = x.SoftDeleteFieldAndZeroValue()
	}

	return func(yield func(*er.OrderedConstraint) bool) {
		for key := range table.Keys() {
			if key.IsNil() {
				continue
			}

			c2 := &er.OrderedConstraint{
				Head: er.Head{
					Name: key.Name(),
				},
				Unique:  key.IsUnique(),
				Primary: key.IsPrimary(),
			}

			keyDef := sqlbuilder.GetKeyDef(key)
			if keyDef != nil {
				c2.Method = keyDef.Method()
			}

			if _, ok := m.(sqltype.WithCreationTime); ok {
				if c2.Unique {
					cols := make([]sqlbuilder.Column, 0)

					for col := range key.Cols() {
						if softDeletedField != "" {
							if col.FieldName() == softDeletedField {
								continue
							}
						}

						cols = append(cols, col)
					}

					if len(cols) == 1 {
						def := sqlbuilder.GetColumnDef(cols[0])
						c.uniques[def.Type.String()] = fmt.Sprintf("%s.%s", table.TableName(), cols[0].Name())
					}
				}
			}

			for col := range key.Cols() {
				cn := er.ConstraintColumnName{
					Name: col.Name(),
				}

				for _, o := range keyDef.FieldNameAndOptions() {
					if name := o.Name(); name == col.Name() || name == col.FieldName() {
						cn.Options = o.Options()
					}
				}

				c2.ColumnNames = append(c2.ColumnNames, cn)
			}

			if !yield(c2) {
				return
			}
		}
	}
}

func (c *collector) mayCollectRuntimeDoc(m any, h *er.Head, fieldNames ...string) {
	if docer, ok := m.(interface {
		RuntimeDoc(names ...string) ([]string, bool)
	}); ok {
		if lines, ok := docer.RuntimeDoc(fieldNames...); ok {
			c.setTitleOrDescription(h, lines)
		}
	}
}

func (c *collector) setTitleOrDescription(head *er.Head, lines []string) {
	if head == nil {
		return
	}

	if len(lines) > 0 {
		head.Title = strings.TrimSpace(lines[0])

		if len(lines) > 1 {
			head.Description = strings.TrimSpace(strings.Join(slices.Collect(filterLine(slices.Values(lines[1:]))), "\n"))
		}
	}
}

func filterLine(seq iter.Seq[string]) iter.Seq[string] {
	return func(yield func(string) bool) {
		for l := range seq {
			if strings.HasPrefix(l, "openapi:") {
				continue
			}

			if !yield(l) {
				return
			}
		}
	}
}
