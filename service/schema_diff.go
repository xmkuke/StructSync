package service

type SchemaDiff struct {
	Table  string
	Source *MySchema
	Dest   *MySchema
}

/*
* Create compare object Ex
 */
func newSchemaDiff(table, dest string, source *MySchema) *SchemaDiff {
	return &SchemaDiff{
		Table:  table,
		Source: source,
		Dest:   ParseSchema(dest),
	}
}

/**
* Get relation tables
 */
func (sDiff *SchemaDiff) RelationTables() []string {
	return sDiff.Source.RelationTables()
}
