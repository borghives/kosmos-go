package expression

import "go.mongodb.org/mongo-driver/v2/bson"

type QueryFieldPredicate struct {
	FieldName FieldName
	Query     QueryOp
}

func (q QueryFieldPredicate) ToRepr() any {
	return bson.D{kv(q.FieldName.Name, q.Query.ToRepr())}
}

func (q QueryFieldPredicate) Reduce(resolver NameResolver) any {
	return bson.D{kv(resolver(q.FieldName.Name), NormalizeExpression(q.Query, resolver))}
}

type QueryOp struct {
	Operator string
	Value    any
}

func (q QueryOp) ToRepr() any {
	return bson.D{kv(q.Operator, q.Value)}
}

func Eq(value any) QueryOp {
	return QueryOp{"$eq", value}
}

func Ne(value any) QueryOp {
	return QueryOp{"$ne", value}
}

func Gt(value any) QueryOp {
	return QueryOp{"$gt", value}
}

func Gte(value any) QueryOp {
	return QueryOp{"$gte", value}
}

func Lt(value any) QueryOp {
	return QueryOp{"$lt", value}
}

func Lte(value any) QueryOp {
	return QueryOp{"$lte", value}
}

func In(values bson.A) QueryOp {
	return QueryOp{"$in", values}
}

func Nin(values bson.A) QueryOp {
	return QueryOp{"$nin", values}
}

func And(values bson.A) QueryOp {
	return QueryOp{"$and", values}
}

// Or returns a query operator that matches documents that satisfy any of the specified conditions.
func Or(values bson.A) QueryOp {
	return QueryOp{"$or", values}
}
