package database

type Condition struct {
	In  map[string]any
	Out map[string]any

	Includes map[string]any
	Excludes map[string]any

	Preload    []string
	Joins      []string
	Where      []Where
	Pagination Paginating
	Sorting    Sorting
}
