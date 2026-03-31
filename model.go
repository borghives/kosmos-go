package kosmos

import (
	"reflect"
	"time"

	"github.com/borghives/kosmos-go/model"
	"github.com/borghives/kosmos-go/model/operator"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type ModelMetadata struct {
	CollectionName string
	DatabaseName   string
}

func (e *ModelMetadata) NormalizeDocument(document bson.D) bson.D {
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

func (e *ModelMetadata) NormalizeArray(array bson.A) bson.A {
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

func (e *ModelMetadata) NormalizeExpression(expression operator.Expression) any {
	rep := expression.ToRepr()
	switch rep := rep.(type) {
	case bson.A:
		return e.NormalizeArray(rep)
	case bson.D:
		return e.NormalizeDocument(rep)
	}
	return rep
}

type Model interface {
	CollapseID() bool
	IsEntangled() bool
	GetMetadata() ModelMetadata
}

type BaseModel struct {
	// KModelMetadata  kosmos.ModelMetadata `xml:"-" json:"-" bson:"-" db:"-" collection:"-"`
	ID          bson.ObjectID `xml:"id,attr" json:"ID" bson:"_id,omitempty"`
	UpdatedTime time.Time     `xml:"updated" json:"updated" bson:"updated_time"`
	CreatedTime time.Time     `xml:"created" json:"created" bson:"created_time"`
}

func (e *BaseModel) CollapseID() {
	if e.ID.IsZero() {
		e.ID = bson.NewObjectID()
		e.CreatedTime = time.Now()
	}
	e.UpdatedTime = time.Now()
}

func (e *BaseModel) IsEntangled() bool {
	return !e.ID.IsZero()
}

func (e *BaseModel) GetMetadata() ModelMetadata {
	field, found := reflect.TypeOf(e).Elem().FieldByName("ModelMetadata")
	if !found {
		panic("ModelMetadata not found")
	}

	return ModelMetadata{
		DatabaseName:   field.Tag.Get("db"),
		CollectionName: field.Tag.Get("collection"),
	}
}

func Filter[T Model](filter model.QueryPredicate) *EntityRecord[T] {
	return All[T]().Filter(filter)
}

func All[T Model]() *EntityRecord[T] {
	var template T
	recording := &EntityRecord[T]{
		Type: template.GetMetadata(),
	}
	return recording
}

func kv(key string, value any) bson.E {
	return bson.E{Key: key, Value: value}
}
