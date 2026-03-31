package operator

import "go.mongodb.org/mongo-driver/v2/bson"

type QueryOpExpression struct {
	Operator string
	Value    any
}

func (q *QueryOpExpression) ToRepr() any {
	return bson.D{kv(q.Operator, q.Value)}
}

func Eq(value any) QueryOpExpression {
	return QueryOpExpression{"$eq", value}
}
