package kmodel

import (
	"reflect"
	"time"

	"github.com/borghives/kosmos-go/kmodel/operator"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Meta struct {
	CollectionName string
	DatabaseName   string
}

func (e Meta) IsValid() bool {
	return e.CollectionName != "" && e.DatabaseName != ""
}

func (e *Meta) NormalizeDocument(document bson.D) bson.D {
	newD := bson.D{}
	for _, v := range document {
		if _, ok := v.Value.(operator.Expression); ok {
			newD = append(newD, kv(v.Key, e.NormalizeExpression(v.Value.(operator.Expression))))
		} else {
			newD = append(newD, v)
		}
	}
	return newD
}

func (e *Meta) NormalizeArray(array bson.A) bson.A {
	newA := bson.A{}
	for _, v := range array {
		if _, ok := v.(operator.Expression); ok {
			newA = append(newA, e.NormalizeExpression(v.(operator.Expression)))
		} else {
			newA = append(newA, v)
		}
	}
	return newA
}

func (e *Meta) NormalizeExpression(expression operator.Expression) any {
	rep := expression.ToRepr()
	switch rep := rep.(type) {
	case bson.A:
		return e.NormalizeArray(rep)
	case bson.D:
		return e.NormalizeDocument(rep)
	}
	return rep
}

type Entity interface {
	CollapseID() bool
	IsEntangled() bool
	GetMeta() Meta
}

type EntityBase struct {
	// EntityModelMeta  entitymodel.Meta `xml:"-" json:"-" bson:"-" db:"-" collection:"-"`
	ID          bson.ObjectID `xml:"id,attr" json:"ID" bson:"_id,omitempty"`
	UpdatedTime time.Time     `xml:"updated" json:"updated" bson:"updated_time"`
	CreatedTime time.Time     `xml:"created" json:"created" bson:"created_time"`
}

func (e *EntityBase) CollapseID() {
	if e.ID.IsZero() {
		e.ID = bson.NewObjectID()
		e.CreatedTime = time.Now()
	}
	e.UpdatedTime = time.Now()
}

func (e *EntityBase) IsEntangled() bool {
	return !e.ID.IsZero()
}

func (e *EntityBase) GetEntityType() Meta {
	field, found := reflect.TypeOf(e).Elem().FieldByName("EntityModelMeta")
	if !found {
		panic("EntityType not found")
	}

	return Meta{
		DatabaseName:   field.Tag.Get("db"),
		CollectionName: field.Tag.Get("collection"),
	}
}

func Filter[T Entity](filter QueryPredicate) *EntityRecord[T] {
	return All[T]().Filter(filter)
}

func All[T Entity]() *EntityRecord[T] {
	var template T
	recording := &EntityRecord[T]{
		Type: template.GetMeta(),
	}
	return recording
}
