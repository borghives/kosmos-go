package kmodel

import (
	"github.com/borghives/kosmos-go/kmodel/operator"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type QueryPredicate struct {
	FieldName  *operator.FieldName
	Expression operator.QueryOpExpression
}

func (q QueryPredicate) ToRepr() any {
	return bson.D{kv(q.FieldName.Name, q.Expression.ToRepr())}
}

// --- Field ---
type QueryableField struct {
	Name string
}

func (q QueryableField) WrapName() *operator.FieldName {
	return &operator.FieldName{Name: q.Name}
}

func (q QueryableField) ToQueryPredicate(queryOp operator.QueryOpExpression) QueryPredicate {
	return QueryPredicate{
		FieldName:  q.WrapName(),
		Expression: queryOp,
	}
}

func (q QueryableField) LiteralSlice(values []any) []operator.LiteralValue {
	literals := make([]operator.LiteralValue, len(values))
	for i, v := range values {
		literals[i] = q.Literal(v)
	}
	return literals
}

func (q QueryableField) Literal(value any) operator.LiteralValue {

	return operator.LiteralValue{
		Value: value,
		Field: q.Name,
	}
}

func (q QueryableField) Eq(value any) QueryPredicate {
	litValue := q.Literal(value)
	return q.ToQueryPredicate(operator.Eq(litValue))
}
