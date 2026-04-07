package observation

import (
	"time"

	"github.com/borghives/kosmos-go/observation/expression"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type EntityField struct {
	Name string
}

type EntityIDField struct {
	EntityField
}

type EntityStringField struct {
	EntityField
}

func (q EntityField) wrapFieldName() expression.FieldName {
	return expression.FieldName{Name: q.Name}
}

func (q EntityField) ToQueryPredicate(queryOp expression.QueryOp) expression.QueryFieldPredicate {
	return expression.QueryFieldPredicate{
		FieldName: q.wrapFieldName(),
		Query:     queryOp,
	}
}

func (q EntityField) literalSlice(values []any) bson.A {
	literals := make(bson.A, len(values))
	for i, v := range values {
		literals[i] = q.literal(v)
	}
	return literals
}

func (q EntityField) literal(value any) expression.LiteralValue {
	switch val := value.(type) {
	case *bson.ObjectID:
		value = *val
	case *string:
		value = *val
	case *int:
		value = *val
	case *int32:
		value = *val
	case *int64:
		value = *val
	case *float32:
		value = *val
	case *float64:
		value = *val
	case *bool:
		value = *val
	case *time.Time:
		value = *val
	}
	return expression.LiteralValue{
		Value:   value,
		Context: q.wrapFieldName(),
	}
}

func (q EntityField) Eq(value any) expression.QueryFieldPredicate {
	if value == nil {
		return expression.QueryFieldPredicate{}
	}
	litValue := q.literal(value)
	return q.ToQueryPredicate(expression.Eq(litValue))
}

func (q EntityField) Ne(value any) expression.QueryFieldPredicate {
	if value == nil {
		return expression.QueryFieldPredicate{}
	}
	return q.ToQueryPredicate(expression.Ne(q.literal(value)))
}

func (q EntityField) Gt(value any) expression.QueryFieldPredicate {
	if value == nil {
		return expression.QueryFieldPredicate{}
	}
	return q.ToQueryPredicate(expression.Gt(q.literal(value)))
}

func (q EntityField) Gte(value any) expression.QueryFieldPredicate {
	if value == nil {
		return expression.QueryFieldPredicate{}
	}
	return q.ToQueryPredicate(expression.Gte(q.literal(value)))
}

func (q EntityField) Lt(value any) expression.QueryFieldPredicate {
	if value == nil {
		return expression.QueryFieldPredicate{}
	}
	return q.ToQueryPredicate(expression.Lt(q.literal(value)))
}

func (q EntityField) Lte(value any) expression.QueryFieldPredicate {
	if value == nil {
		return expression.QueryFieldPredicate{}
	}
	return q.ToQueryPredicate(expression.Lte(q.literal(value)))
}

func (q EntityField) In(values ...any) expression.QueryFieldPredicate {
	if values == nil {
		return expression.QueryFieldPredicate{}
	}
	return q.ToQueryPredicate(expression.In(q.literalSlice(values)))
}

func (q EntityField) Nin(values ...any) expression.QueryFieldPredicate {
	if values == nil {
		return expression.QueryFieldPredicate{}
	}
	return q.ToQueryPredicate(expression.Nin(q.literalSlice(values)))
}

func (q EntityField) ID() EntityIDField {
	return EntityIDField{EntityField: q}
}

func (q EntityField) Str() EntityStringField {
	return EntityStringField{EntityField: q}
}

// --- Entity ID Field ---
func (q EntityIDField) In(values ...bson.ObjectID) expression.QueryFieldPredicate {
	return q.ToQueryPredicate(expression.In(q.literalSlice(values)))
}

func (q EntityIDField) Nin(values ...bson.ObjectID) expression.QueryFieldPredicate {
	return q.ToQueryPredicate(expression.Nin(q.literalSlice(values)))
}

func (q EntityIDField) literalSlice(values []bson.ObjectID) bson.A {
	literals := make(bson.A, len(values))
	for i, v := range values {
		literals[i] = q.literal(v)
	}
	return literals
}

// --- Entity String Field ---

func (q EntityStringField) In(values ...string) expression.QueryFieldPredicate {
	return q.ToQueryPredicate(expression.In(q.literalSlice(values)))
}

func (q EntityStringField) Nin(values ...string) expression.QueryFieldPredicate {
	return q.ToQueryPredicate(expression.Nin(q.literalSlice(values)))
}

func (q EntityStringField) literalSlice(values []string) bson.A {
	literals := make(bson.A, len(values))
	for i, v := range values {
		literals[i] = q.literal(v)
	}
	return literals
}
