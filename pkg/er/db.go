// Package er 定义实体关系结构及其保序表示。
package er

// Head 表示 ER 节点的通用元信息。
type Head struct {
	Name string `json:"-"`

	Title       string `json:"title,omitzero"`
	Description string `json:"description,omitzero"`
}

// Database 表示 ER 数据库结构。
type Database struct {
	Head

	Tables map[string]*Table `json:"tables"`
}

// Table 表示 ER 表结构。
type Table struct {
	Head

	Columns     map[string]*Column     `json:"columns"`
	Constraints map[string]*Constraint `json:"constraints"`
}

// Column 表示 ER 列结构。
type Column struct {
	Head

	Type string `json:"type"`
	Of   string `json:"of,omitzero"`

	GoType string `json:"-"`
}

// Constraint 表示 ER 约束结构。
type Constraint struct {
	Head

	ColumnNames []ConstraintColumnName `json:"columnNames"`
	Method      string                 `json:"method,omitzero"`
	Unique      bool                   `json:"unique,omitzero"`
	Primary     bool                   `json:"primary,omitzero"`
}

// ConstraintColumnName 表示约束中的列及附加选项。
type ConstraintColumnName struct {
	Name    string   `json:"name"`
	Options []string `json:"options,omitzero"`
}
