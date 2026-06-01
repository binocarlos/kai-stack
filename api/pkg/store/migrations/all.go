package migrations

// All is the ordered list of migrations applied to the database, in order.
//
// IMPORTANT: only ever APPEND new migrations to the end of this slice, and
// never rename or reorder existing entries. The Name is recorded in the
// go_schema_migrations ledger once applied; reordering or renaming would make
// the runner re-apply or skip the wrong migrations.
var All = []Migration{
	{Name: "000001_initial_schema", Up: InitialSchema},
	{Name: "000002_vector_support", Up: VectorSupport},
	{Name: "000003_jobs", Up: Jobs},
}
