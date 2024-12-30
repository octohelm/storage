package extractor

import (
	"context"
	"fmt"
	"iter"
	"slices"
	"strings"

	"github.com/octohelm/storage/pkg/er"
	"github.com/octohelm/storage/pkg/session"
	"github.com/octohelm/storage/pkg/sqlbuilder"
	"github.com/octohelm/storage/pkg/sqlfrag"
)

func FromCatalog(ctx context.Context, s session.Session, tables *sqlbuilder.Tables) *er.Database {
	erd := &er.Database{
		Name:   s.Name(),
		Tables: make(map[string]*er.Table),
	}

	c := &collector{
		s:         s,
		primaries: make(map[string]string),
	}

	for t := range c.tables(ctx, tables) {
		erd.Tables[t.Name] = t
	}

	for _, t := range erd.Tables {
		for _, col := range t.Columns {
			if of, ok := c.primaries[col.GoType]; ok {
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

	primaries map[string]string
}

func (c *collector) tables(ctx context.Context, tables *sqlbuilder.Tables) iter.Seq[*er.Table] {
	return func(yield func(*er.Table) bool) {
		for table := range tables.Range {
			t := &er.Table{
				Head: er.Head{
					Name: table.TableName(),
				},
				Columns:     make(map[string]*er.Column),
				Constraints: make(map[string]*er.Constraint),
			}

			v, ok := table.(interface{ New() sqlbuilder.Model })
			if !ok {
				continue
			}
			m := v.New()

			for col := range c.columns(ctx, table, m) {
				t.Columns[col.Name] = col
			}

			for cc := range c.constraints(ctx, table, m) {
				t.Constraints[cc.Name] = cc
			}

			c.mayCollectRuntimeDoc(m, &t.Head)

			if !yield(t) {
				return
			}
		}
	}
}

func (c *collector) columns(ctx context.Context, table sqlbuilder.Table, m sqlbuilder.Model) iter.Seq[*er.Column] {
	return func(yield func(*er.Column) bool) {
		for col := range table.Cols() {
			def := sqlbuilder.GetColumnDef(col)
			if def.DeprecatedActions != nil {
				continue
			}

			c2 := &er.Column{
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

func (c *collector) constraints(ctx context.Context, table sqlbuilder.Table, m sqlbuilder.Model) iter.Seq[*er.Constraint] {
	return func(yield func(*er.Constraint) bool) {
		for key := range table.Keys() {
			if key.IsNil() {
				continue
			}

			c2 := &er.Constraint{
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

			for col := range key.Cols() {
				if c2.Primary {
					if strings.Contains(col.Name(), "f_id") {
						def := sqlbuilder.GetColumnDef(col)
						c.primaries[def.Type.String()] = fmt.Sprintf("%s.%s", table.TableName(), col.Name())
					}
				}

				c2.ColumnNames = append(c2.ColumnNames, col.Name())
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
