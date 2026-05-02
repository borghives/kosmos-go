package expression

import (
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Base interface {
	ToRepr() any
}

type Reducible interface {
	Reduce(resolver NameResolver) any
}

type FieldName struct {
	Name string
}

func (f FieldName) ToRepr() any {
	return f.Name
}

func (f FieldName) Reduce(resolver NameResolver) any {
	return resolver(f.Name)
}

type LiteralValue struct {
	Value   any
	Context FieldName
}

func (l LiteralValue) ToRepr() any {
	return l.Value
}

func (l LiteralValue) Reduce(resolver NameResolver) any {
	return l.Value
}

func kv(key string, value any) bson.E {
	return bson.E{Key: key, Value: value}
}

// --- Normalize Expression ---
type NameResolver func(string) string

func NormalizeExpression(expr any, resolver NameResolver) any {
	if expr, ok := expr.(Reducible); ok {
		return expr.Reduce(resolver)
	}

	switch rep := expr.(type) {
	case bson.A:
		return NormalizeArray(rep, resolver)
	case bson.D:
		return NormalizeDocument(rep, resolver)
	case Base:
		return NormalizeExpression(rep.ToRepr(), resolver)
	default:
		return rep
	}
}

func NormalizeDocument(document bson.D, resolver NameResolver) bson.D {
	newD := make(bson.D, 0, len(document))
	for _, v := range document {
		switch val := v.Value.(type) {
		case bson.D:
			newD = append(newD, kv(v.Key, NormalizeDocument(val, resolver)))
		case bson.A:
			newD = append(newD, kv(v.Key, NormalizeArray(val, resolver)))
		default:
			newD = append(newD, kv(v.Key, NormalizeExpression(val, resolver)))
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
