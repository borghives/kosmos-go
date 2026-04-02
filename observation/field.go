package observation

import (
	"github.com/borghives/kosmos-go/observation/expression"
	"go.mongodb.org/mongo-driver/v2/bson"
)

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

func (q QueryField) literalSlice(values []any) bson.A {
	literals := make(bson.A, len(values))
	for i, v := range values {
		literals[i] = q.literal(v)
	}
	return literals
}

func (q QueryField) literal(value any) expression.LiteralValue {
	return expression.LiteralValue{
		Value:   value,
		Context: q.wrapFieldName(),
	}
}

func (q QueryField) Eq(value any) expression.QueryFieldPredicate {
	litValue := q.literal(value)
	return q.ToQueryPredicate(expression.Eq(litValue))
}

func (q QueryField) Ne(value any) expression.QueryFieldPredicate {
	return q.ToQueryPredicate(expression.Ne(q.literal(value)))
}

func (q QueryField) Gt(value any) expression.QueryFieldPredicate {
	return q.ToQueryPredicate(expression.Gt(q.literal(value)))
}

func (q QueryField) Gte(value any) expression.QueryFieldPredicate {
	return q.ToQueryPredicate(expression.Gte(q.literal(value)))
}

func (q QueryField) Lt(value any) expression.QueryFieldPredicate {
	return q.ToQueryPredicate(expression.Lt(q.literal(value)))
}

func (q QueryField) Lte(value any) expression.QueryFieldPredicate {
	return q.ToQueryPredicate(expression.Lte(q.literal(value)))
}

func (q QueryField) In(values ...any) expression.QueryFieldPredicate {
	return q.ToQueryPredicate(expression.In(q.literalSlice(values)))
}

func (q QueryField) Nin(values ...any) expression.QueryFieldPredicate {
	return q.ToQueryPredicate(expression.Nin(q.literalSlice(values)))
}
