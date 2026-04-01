package expression

import "go.mongodb.org/mongo-driver/v2/bson"

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

func In(value any) QueryOp {
	return QueryOp{"$in", value}
}

func Nin(value any) QueryOp {
	return QueryOp{"$nin", value}
}

func And(values ...Base) QueryOp {
	valuesA := make(bson.A, len(values))
	for i, v := range values {
		valuesA[i] = v
	}
	return QueryOp{"$and", valuesA}
}

func Or(values ...Base) QueryOp {
	valuesA := make(bson.A, len(values))
	for i, v := range values {
		valuesA[i] = v
	}
	return QueryOp{"$or", valuesA}
}
