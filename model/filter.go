package model

import (
	"github.com/borghives/kosmos-go/model/expression"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type QueryPredicate struct {
	FieldName *expression.FieldName
	Query     expression.QueryOp
}

func (q QueryPredicate) ToRepr() any {
	return bson.D{kv(q.FieldName.Name, q.Query.ToRepr())}
}

// --- QueryField ---
func Fld(name string) QueryField {
	return QueryField{Name: name}
}

type QueryField struct {
	Name string
}

func (q QueryField) wrapFieldName() *expression.FieldName {
	return &expression.FieldName{Name: q.Name}
}

func (q QueryField) ToQueryPredicate(queryOp expression.QueryOp) QueryPredicate {
	return QueryPredicate{
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

func (q QueryField) Eq(value any) QueryPredicate {
	litValue := q.Literal(value)
	return q.ToQueryPredicate(expression.Eq(litValue))
}

func (q QueryField) Ne(value any) QueryPredicate {
	return q.ToQueryPredicate(expression.Ne(q.Literal(value)))
}

func (q QueryField) Gt(value any) QueryPredicate {
	return q.ToQueryPredicate(expression.Gt(q.Literal(value)))
}

func (q QueryField) Gte(value any) QueryPredicate {
	return q.ToQueryPredicate(expression.Gte(q.Literal(value)))
}

func (q QueryField) Lt(value any) QueryPredicate {
	return q.ToQueryPredicate(expression.Lt(q.Literal(value)))
}

func (q QueryField) Lte(value any) QueryPredicate {
	return q.ToQueryPredicate(expression.Lte(q.Literal(value)))
}

func (q QueryField) In(values ...any) QueryPredicate {
	return q.ToQueryPredicate(expression.In(q.LiteralSlice(values)))
}

func (q QueryField) Nin(values ...any) QueryPredicate {
	return q.ToQueryPredicate(expression.Nin(q.LiteralSlice(values)))
}
