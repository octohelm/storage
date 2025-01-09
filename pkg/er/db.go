package er

type Head struct {
	Name string `json:"-"`

	Title       string `json:"title,omitzero"`
	Description string `json:"description,omitzero"`
}

type Database struct {
	Head

	Tables map[string]*Table `json:"tables"`
}

type Table struct {
	Head

	Columns     map[string]*Column     `json:"columns"`
	Constraints map[string]*Constraint `json:"constraints"`
}

type Column struct {
	Head

	Type string `json:"type"`
	Of   string `json:"of,omitzero"`

	GoType string `json:"-"`
}

type Constraint struct {
	Head

	ColumnNames []ConstraintColumnName `json:"columnNames"`
	Method      string                 `json:"method,omitzero"`
	Unique      bool                   `json:"unique,omitzero"`
	Primary     bool                   `json:"primary,omitzero"`
}

type ConstraintColumnName struct {
	Name    string   `json:"name"`
	Options []string `json:"options,omitzero"`
}
