package sqlbuilder

import (
	"container/list"
	"iter"
)

type Catalog interface {
	Table(tableName string) Table
	Tables() iter.Seq[Table]

	Add(tabs ...Table)
	Remove(tableName string)

	Require(catalogs ...Catalog)
}

func TableNames(c Catalog) iter.Seq[string] {
	return func(yield func(string) bool) {
		for t := range c.Tables() {
			if !yield(t.TableName()) {
				return
			}
		}
	}
}

var _ Catalog = &Tables{}

type Tables struct {
	l      *list.List
	tables map[string]*list.Element

	requirements []Catalog
}

func (c *Tables) Require(catalogs ...Catalog) {
	c.requirements = append(c.requirements, catalogs...)
}

func (c *Tables) Add(tabs ...Table) {
	if c.tables == nil {
		c.tables = map[string]*list.Element{}
		c.l = list.New()
	}

	for _, tab := range tabs {
		if tab != nil {
			if _, ok := c.tables[tab.TableName()]; ok {
				c.Remove(tab.TableName())
			}
			e := c.l.PushBack(tab)
			c.tables[tab.TableName()] = e
		}
	}
}

func (c *Tables) Table(tableName string) Table {
	if c.tables != nil {
		if c, ok := c.tables[tableName]; ok {
			return c.Value.(Table)
		}
	}
	return nil
}

func (c *Tables) Remove(name string) {
	if c.tables != nil {
		if e, exists := c.tables[name]; exists {
			c.l.Remove(e)
			delete(c.tables, name)
		}
	}
}

func (c Tables) Tables() iter.Seq[Table] {

	return func(yield func(Table) bool) {
		emitted := make(map[string]bool)

		emit := func(t Table) bool {
			tableName := t.TableName()

			if _, ok := emitted[tableName]; ok {
				return true
			}

			emitted[tableName] = true
			return yield(t)
		}

		if c.l != nil {
			for e := c.l.Front(); e != nil; e = e.Next() {
				t := e.Value.(Table)

				if !emit(t) {
					return
				}
			}
		}

		for _, cc := range c.requirements {
			for t := range cc.Tables() {
				if !emit(t) {
					return
				}
			}
		}
	}
}
