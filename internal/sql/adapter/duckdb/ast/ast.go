package ast

import (
	duckdbtoken "github.com/octohelm/storage/internal/sql/adapter/duckdb/token"
)

type Expr interface {
	exprNode()
}

type Ident struct {
	Name   string
	Quoted bool
}

type BasicLit struct {
	Kind  duckdbtoken.Token
	Value string
}

type TypeAssertExpr struct {
	X    Expr
	Type Expr
}

type CallExpr struct {
	Fun  Expr
	Args []Expr
}

type BinaryExpr struct {
	Op duckdbtoken.Token
	X  Expr
	Y  Expr
}

type UnaryExpr struct {
	Op duckdbtoken.Token
	X  Expr
}

func (Ident) exprNode()          {}
func (BasicLit) exprNode()       {}
func (TypeAssertExpr) exprNode() {}
func (CallExpr) exprNode()       {}
func (BinaryExpr) exprNode()     {}
func (UnaryExpr) exprNode()      {}

type SelectorExpr struct {
	X   Expr
	Sel *Ident
}

func (SelectorExpr) exprNode() {}

type IndexExpr struct {
	X   Expr
	Sel *Ident
}

func (IndexExpr) exprNode() {}

type Stmt interface {
	stmtNode()
}

type CreateStmt struct {
	Name        Expr
	Columns     []ColumnDecl
	Constraints []TableConstraint
	OrReplace   bool
	Temporary   bool
	IfNotExists bool
}

func (CreateStmt) stmtNode() {}

type ColumnDecl interface {
	ColumnName() string
}

type ColumnDeclBasic struct {
	Name       *Ident
	Type       *Ident
	Default    Expr
	Check      Expr
	PrimaryKey bool
	NotNull    bool
	Unique     bool
	Collate    *Ident
}

func (c *ColumnDeclBasic) ColumnName() string {
	return c.Name.Name
}

type ColumnDeclGenerated struct {
	Name   *Ident
	Type   *Ident
	AS     Expr
	Stored bool // false means VIRTUAL
}

func (c *ColumnDeclGenerated) ColumnName() string {
	return c.Name.Name
}

type TableConstraint interface {
	tableConstraint()
}

type TableConstraintPrimaryKey []*Ident

type TableConstraintUnique []*Ident

type TableConstraintForeignKey struct {
	Table   Expr
	Columns []*Ident
}

func (TableConstraintPrimaryKey) tableConstraint() {}
func (TableConstraintUnique) tableConstraint()     {}
func (TableConstraintForeignKey) tableConstraint() {}
