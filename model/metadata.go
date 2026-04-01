package model

import (
	"log"
	"reflect"
	"slices"
	"strings"

	"github.com/borghives/kosmos-go/model/operator"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Metadata struct {
	DatabaseName   string
	CollectionName string
	FieldMap       map[string]string
}

func GetMetadata(obj any) Metadata {
	t := reflect.TypeOf(obj)
	if t == nil {
		log.Fatal("model.GetMetadata: cannot extract metadata from a nil interface")
	}
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	field, found := t.FieldByName("BaseModel")
	if !found {
		log.Fatal("model.GetMetadata: BaseModel not found")
	}

	fieldMap := make(map[string]string)
	populateFieldMap(t, fieldMap)

	return Metadata{
		DatabaseName:   field.Tag.Get("kdb"),
		CollectionName: field.Tag.Get("kcol"),
		FieldMap:       fieldMap,
	}
}

func populateFieldMap(t reflect.Type, m map[string]string) {
	if t.Kind() != reflect.Struct {
		return
	}
	for field := range t.Fields() {
		if !field.IsExported() {
			continue
		}

		bsonTag := field.Tag.Get("bson")
		if bsonTag == "-" {
			continue
		}

		bsonName := field.Name
		inline := false
		if bsonTag != "" {
			parts := strings.Split(bsonTag, ",")
			if parts[0] != "" {
				bsonName = parts[0]
			}
			inline = slices.Contains(parts[1:], "inline")
		}

		m[field.Name] = bsonName

		if inline || field.Anonymous {
			fieldType := field.Type
			for fieldType.Kind() == reflect.Pointer {
				fieldType = fieldType.Elem()
			}
			if fieldType.Kind() == reflect.Struct {
				populateFieldMap(fieldType, m)
			}
		}
	}
}

func (e *Metadata) resolveAlias(name string) string {
	if e.FieldMap != nil {
		if mapped, ok := e.FieldMap[name]; ok {
			return mapped
		}
	}
	return name
}

func (e *Metadata) NormalizeDocument(document bson.D) bson.D {
	newD := make(bson.D, 0, len(document))
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
	newA := make(bson.A, 0, len(array))
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
	switch expr := expression.(type) {
	case QueryPredicate:
		return e.NormalizeDocument(bson.D{{Key: e.resolveAlias(expr.FieldName.Name), Value: e.NormalizeExpression(expr.Expression)}})
	case *operator.FieldName:
		return e.resolveAlias(expr.Name)
	case operator.FieldName:
		return e.resolveAlias(expr.Name)
	}

	rep := expression.ToRepr()
	switch rep := rep.(type) {
	case bson.A:
		return e.NormalizeArray(rep)
	case bson.D:
		return e.NormalizeDocument(rep)
	}
	return rep
}
