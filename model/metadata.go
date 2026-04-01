package model

import (
	"reflect"

	"github.com/borghives/kosmos-go/model/operator"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Metadata struct {
	DatabaseName   string
	CollectionName string
}

func GetMetadata(obj any) Metadata {
	t := reflect.TypeOf(obj)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	field, found := t.FieldByName("BaseModel")
	if !found {
		panic("BaseModel not found")
	}
	return Metadata{
		DatabaseName:   field.Tag.Get("kdb"),
		CollectionName: field.Tag.Get("kcol"),
	}
}

func (e *Metadata) NormalizeDocument(document bson.D) bson.D {
	newD := bson.D{}
	for _, v := range document {
		switch val := v.Value.(type) {
		case operator.Expression:
			newD = append(newD, kv(v.Key, e.NormalizeExpression(val)))
		case bson.D:
			newD = append(newD, kv(v.Key, e.NormalizeDocument(val)))
		case bson.A:
			newD = append(newD, kv(v.Key, e.NormalizeArray(val)))
		default:
			newD = append(newD, v)
		}
	}
	return newD
}

func (e *Metadata) NormalizeArray(array bson.A) bson.A {
	newA := bson.A{}
	for _, v := range array {
		switch val := v.(type) {
		case operator.Expression:
			newA = append(newA, e.NormalizeExpression(val))
		case bson.D:
			newA = append(newA, e.NormalizeDocument(val))
		case bson.A:
			newA = append(newA, e.NormalizeArray(val))
		default:
			newA = append(newA, v)
		}
	}
	return newA
}

func (e *Metadata) NormalizeExpression(expression operator.Expression) any {
	rep := expression.ToRepr()
	switch rep := rep.(type) {
	case bson.A:
		return e.NormalizeArray(rep)
	case bson.D:
		return e.NormalizeDocument(rep)
	}
	return rep
}
