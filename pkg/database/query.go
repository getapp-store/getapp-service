package database

import "gorm.io/gorm"

type Sorting struct {
	Sort  string
	Order string
}

type Paginating struct {
	Start int
	End   int
}

type Param struct {
	Field string
	Value any
}

type Where struct {
	Condition string
	Values    []any
}

type Builder struct {
	Includes   []Param
	Excludes   []Param
	Sorting    Sorting
	Pagination Paginating
	Joins      []string
	Select     []string
	Preload    []string
	Group      []string
	Where      []Where
	In         []Param
	Out        []Param
}

func Query(condition Condition) Builder {
	builder := Builder{}
	builder.Where = condition.Where
	builder.Pagination = condition.Pagination
	builder.Sorting = condition.Sorting
	builder.Preload = condition.Preload
	builder.Joins = condition.Joins

	for field, val := range condition.In {
		builder.In = append(builder.In, Param{
			Field: field,
			Value: val,
		})
	}

	for field, val := range condition.Out {
		builder.Out = append(builder.In, Param{
			Field: field,
			Value: val,
		})
	}

	for field, val := range condition.Includes {
		builder.Includes = append(builder.Includes, Param{
			Field: field,
			Value: val,
		})
	}

	for field, val := range condition.Excludes {
		builder.Excludes = append(builder.Excludes, Param{
			Field: field,
			Value: val,
		})
	}

	return builder
}

func (c Builder) Build(q *gorm.DB) *gorm.DB {
	for _, sel := range c.Select {
		q = q.Select(sel)
	}
	for _, join := range c.Joins {
		q = q.Joins(join)
	}

	for _, preload := range c.Preload {
		q = q.Preload(preload)
	}

	for _, param := range c.Includes {
		q = q.Where(param.Field+" = ?", param.Value)
	}

	for _, param := range c.Excludes {
		q = q.Where(param.Field+" != ?", param.Value)
	}

	for _, param := range c.In {
		q = q.Where(param.Field+" in (?)", param.Value)
	}

	for _, param := range c.Out {
		q = q.Where(param.Field+" not in (?)", param.Value)
	}

	for _, where := range c.Where {
		//expr := gorm.Expr(where.Condition, where.Values...)
		q = q.Where(where.Condition, where.Values...)
	}

	for _, column := range c.Group {
		q = q.Group(column)
	}

	sort := c.Sorting.Sort
	order := c.Sorting.Order

	if sort != "" && order != "" {
		q = q.Order(sort + " " + order)
	}

	start := c.Pagination.Start
	end := c.Pagination.End

	if end > start {
		q = q.Offset(start).Limit(end - start)
	}

	return q
}
