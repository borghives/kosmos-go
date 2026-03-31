package model

import (
	"github.com/borghives/kosmos-go/model/operator"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type QueryPredicate struct {
	FieldName  *operator.FieldName
	Expression operator.QueryOpExpression
}

func (q QueryPredicate) ToRepr() any {
	return bson.D{kv(q.FieldName.Name, q.Expression.ToRepr())}
}

// --- QueryField ---
func Fld(name string) QueryField {
	return QueryField{Name: name}
}

type QueryField struct {
	Name string
}

func (q QueryField) wrapFieldName() *operator.FieldName {
	return &operator.FieldName{Name: q.Name}
}

func (q QueryField) ToQueryPredicate(queryOp operator.QueryOpExpression) QueryPredicate {
	return QueryPredicate{
		FieldName:  q.wrapFieldName(),
		Expression: queryOp,
	}
}

func (q QueryField) LiteralSlice(values []any) []operator.LiteralValue {
	literals := make([]operator.LiteralValue, len(values))
	for i, v := range values {
		literals[i] = q.Literal(v)
	}
	return literals
}

func (q QueryField) Literal(value any) operator.LiteralValue {

	return operator.LiteralValue{
		Value: value,
		Field: q.Name,
	}
}

func (q QueryField) Eq(value any) QueryPredicate {
	litValue := q.Literal(value)
	return q.ToQueryPredicate(operator.Eq(litValue))
}
