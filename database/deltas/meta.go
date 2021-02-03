package deltas

import "github.com/beaconsoftwarellc/gadget/database/qb"

type deltaMeta struct {
	alias    string
	ID       qb.TableField
	Name     qb.TableField
	Created  qb.TableField
	Modified qb.TableField

	allColumns qb.TableField
}

func (p *deltaMeta) AllColumns() qb.TableField {
	return p.allColumns
}

func (p *deltaMeta) GetName() string {
	return "delta"
}

func (p *deltaMeta) GetAlias() string {
	return p.alias
}

func (p *deltaMeta) PrimaryKey() qb.TableField {
	return p.ID
}

func (p *deltaMeta) SortBy() (qb.TableField, qb.OrderDirection) {
	return p.Created, qb.Ascending
}

func (p *deltaMeta) ReadColumns() []qb.TableField {
	return []qb.TableField{
		p.ID,
		p.Name,
		p.Created,
		p.Modified,
	}
}

func (p *deltaMeta) WriteColumns() []qb.TableField {
	return []qb.TableField{
		p.ID,
		p.Name,
	}
}

func (p *deltaMeta) Alias(alias string) *deltaMeta {
	return &deltaMeta{
		alias:    alias,
		ID:       qb.TableField{Name: "id", Table: alias},
		Name:     qb.TableField{Name: "name", Table: alias},
		Created:  qb.TableField{Name: "created", Table: alias},
		Modified: qb.TableField{Name: "modified", Table: alias},

		allColumns: qb.TableField{Name: "*", Table: alias},
	}
}

// DeltaMeta is a meta representation of the database delta table for building ad hoc queries
var DeltaMeta = (&deltaMeta{}).Alias("delta")
