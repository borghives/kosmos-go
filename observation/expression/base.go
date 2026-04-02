package expression

import (
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Base interface {
	ToRepr() any
}

type FieldName struct {
	Name string
}

func (f FieldName) ToRepr() any {
	return f.Name
}

type LiteralValue struct {
	Value any
	Field string
}

func (l LiteralValue) ToRepr() any {
	return l.Value
}

func kv(key string, value any) bson.E {
	return bson.E{Key: key, Value: value}
}

// --- Normalize Expression ---
type NameResolver func(string) string

func NormalizeExpression(expr Base, resolver NameResolver) any {
	switch exprType := expr.(type) {
	case QueryFieldPredicate:
		return NormalizeDocument(bson.D{{Key: resolver(exprType.FieldName.Name), Value: NormalizeExpression(exprType.Query, resolver)}}, resolver)
	case FieldName:
		return resolver(exprType.Name)
	}

	rep := expr.ToRepr()
	switch rep := rep.(type) {
	case bson.A:
		return NormalizeArray(rep, resolver)
	case bson.D:
		return NormalizeDocument(rep, resolver)
	}
	return rep
}

func NormalizeDocument(document bson.D, resolver NameResolver) bson.D {
	newD := make(bson.D, 0, len(document))
	for _, v := range document {
		switch val := v.Value.(type) {
		case Base:
			newD = append(newD, kv(v.Key, NormalizeExpression(val, resolver)))
		case bson.D:
			newD = append(newD, kv(v.Key, NormalizeDocument(val, resolver)))
		case bson.A:
			newD = append(newD, kv(v.Key, NormalizeArray(val, resolver)))
		default:
			newD = append(newD, v)
		}
	}
	return newD
}

func NormalizeArray(array bson.A, resolver NameResolver) bson.A {
	newA := make(bson.A, 0, len(array))
	for _, v := range array {
		switch val := v.(type) {
		case Base:
			newA = append(newA, NormalizeExpression(val, resolver))
		case bson.D:
			newA = append(newA, NormalizeDocument(val, resolver))
		case bson.A:
			newA = append(newA, NormalizeArray(val, resolver))
		default:
			newA = append(newA, v)
		}
	}
	return newA
}
