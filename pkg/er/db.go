package er

type Database struct {
	Name   string            `json:"name"`
	Tables map[string]*Table `json:"tables"`
}

func (d *Database) Er() *Database {
	return d
}

type Head struct {
	Name string `json:"-"`

	Title       string `json:"title,omitzero"`
	Description string `json:"description,omitzero"`
}

type Table struct {
	Head

	Columns     map[string]*Column     `json:"columns"`
	Constraints map[string]*Constraint `json:"constraints"`
}

type Column struct {
	Head
	Type   string `json:"type"`
	GoType string `json:"-"`
	Of     string `json:"of,omitzero"`
}

type Constraint struct {
	Head

	ColumnNames []string `json:"columnNames"`
	Method      string   `json:"method,omitzero"`
	Unique      bool     `json:"unique,omitzero"`
	Primary     bool     `json:"primary,omitzero"`
}
