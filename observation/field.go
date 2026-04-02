package observation

import (
	"github.com/borghives/kosmos-go/observation/expression"
)

func Fld(name string) QueryField {
	return QueryField{Name: name}
}

type QueryField struct {
	Name string
}

func (q QueryField) wrapFieldName() expression.FieldName {
	return expression.FieldName{Name: q.Name}
}

func (q QueryField) ToQueryPredicate(queryOp expression.QueryOp) expression.QueryFieldPredicate {
	return expression.QueryFieldPredicate{
		FieldName: q.wrapFieldName(),
		Query:     queryOp,
	}
}

func (q QueryField) LiteralSlice(values []any) []expression.LiteralValue {
	literals := make([]expression.LiteralValue, len(values))
	for i, v := range values {
		literals[i] = q.Literal(v)
	}
	return literals
}

func (q QueryField) Literal(value any) expression.LiteralValue {
	return expression.LiteralValue{
		Value: value,
		Field: q.Name,
	}
}

func (q QueryField) Eq(value any) expression.QueryFieldPredicate {
	litValue := q.Literal(value)
	return q.ToQueryPredicate(expression.Eq(litValue))
}

func (q QueryField) Ne(value any) expression.QueryFieldPredicate {
	return q.ToQueryPredicate(expression.Ne(q.Literal(value)))
}

func (q QueryField) Gt(value any) expression.QueryFieldPredicate {
	return q.ToQueryPredicate(expression.Gt(q.Literal(value)))
}

func (q QueryField) Gte(value any) expression.QueryFieldPredicate {
	return q.ToQueryPredicate(expression.Gte(q.Literal(value)))
}

func (q QueryField) Lt(value any) expression.QueryFieldPredicate {
	return q.ToQueryPredicate(expression.Lt(q.Literal(value)))
}

func (q QueryField) Lte(value any) expression.QueryFieldPredicate {
	return q.ToQueryPredicate(expression.Lte(q.Literal(value)))
}

func (q QueryField) In(values ...any) expression.QueryFieldPredicate {
	return q.ToQueryPredicate(expression.In(q.LiteralSlice(values)))
}

func (q QueryField) Nin(values ...any) expression.QueryFieldPredicate {
	return q.ToQueryPredicate(expression.Nin(q.LiteralSlice(values)))
}
