package operator

import "go.mongodb.org/mongo-driver/v2/bson"

type Expression interface {
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
