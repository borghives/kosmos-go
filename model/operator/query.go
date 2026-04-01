package operator

import "go.mongodb.org/mongo-driver/v2/bson"

type QueryOpExpression struct {
	Operator string
	Value    any
}

func (q QueryOpExpression) ToRepr() any {
	return bson.D{kv(q.Operator, q.Value)}
}

func Eq(value any) QueryOpExpression {
	return QueryOpExpression{"$eq", value}
}

func Ne(value any) QueryOpExpression {
	return QueryOpExpression{"$ne", value}
}

func Gt(value any) QueryOpExpression {
	return QueryOpExpression{"$gt", value}
}

func Gte(value any) QueryOpExpression {
	return QueryOpExpression{"$gte", value}
}

func Lt(value any) QueryOpExpression {
	return QueryOpExpression{"$lt", value}
}

func Lte(value any) QueryOpExpression {
	return QueryOpExpression{"$lte", value}
}

func In(value any) QueryOpExpression {
	return QueryOpExpression{"$in", value}
}

func Nin(value any) QueryOpExpression {
	return QueryOpExpression{"$nin", value}
}
