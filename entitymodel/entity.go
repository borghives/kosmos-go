package entitymodel

import (
	"reflect"
	"time"

	"github.com/borghives/kosmos-go/entitymodel/operator"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type EntityObservation struct {
	CollectionName string
	DatabaseName   string
}

func (e EntityObservation) IsValid() bool {
	return e.CollectionName != "" && e.DatabaseName != ""
}

func (e *EntityObservation) NormalizeDocument(document bson.D) bson.D {
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

func (e *EntityObservation) NormalizeArray(array bson.A) bson.A {
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

func (e *EntityObservation) NormalizeExpression(expression operator.Expression) any {
	rep := expression.ToRepr()
	switch rep := rep.(type) {
	case bson.A:
		return e.NormalizeArray(rep)
	case bson.D:
		return e.NormalizeDocument(rep)
	}
	return rep
}

type EntityModel interface {
	CollapseID() bool
	IsEntangled() bool
	GetEntityType() EntityObservation
}

type EntityModelBase struct {
	// EntityObserver  EntityObservation `xml:"-" json:"-" bson:"-" db:"-" collection:"-"`
	ID          bson.ObjectID `xml:"id,attr" json:"ID" bson:"_id,omitempty"`
	UpdatedTime time.Time     `xml:"updated" json:"updated" bson:"updated_time"`
	CreatedTime time.Time     `xml:"created" json:"created" bson:"created_time"`
}

func (e *EntityModelBase) CollapseID() {
	if e.ID.IsZero() {
		e.ID = bson.NewObjectID()
		e.CreatedTime = time.Now()
	}
	e.UpdatedTime = time.Now()
}

func (e *EntityModelBase) IsEntangled() bool {
	return !e.ID.IsZero()
}

func (e *EntityModelBase) GetEntityType() EntityObservation {
	field, found := reflect.TypeOf(e).Elem().FieldByName("EntityObserver")
	if !found {
		panic("EntityObserver not found")
	}

	return EntityObservation{
		DatabaseName:   field.Tag.Get("db"),
		CollectionName: field.Tag.Get("collection"),
	}
}

func Filter[T EntityModel](filter QueryPredicate) *EntityRecord[T] {
	var template T
	recording := &EntityRecord[T]{
		Type: template.GetEntityType(),
	}
	return recording.Filter(filter)
}

func All[T EntityModel]() *EntityRecord[T] {
	var template T
	recording := &EntityRecord[T]{
		Type: template.GetEntityType(),
	}
	return recording
}
