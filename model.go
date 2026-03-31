package kosmos

import (
	"reflect"
	"time"

	"github.com/borghives/kosmos-go/model"
	"github.com/borghives/kosmos-go/model/operator"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type ModelMeta struct {
	DatabaseName   string
	CollectionName string
}

func (e *ModelMeta) NormalizeDocument(document bson.D) bson.D {
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

func (e *ModelMeta) NormalizeArray(array bson.A) bson.A {
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

func (e *ModelMeta) NormalizeExpression(expression operator.Expression) any {
	rep := expression.ToRepr()
	switch rep := rep.(type) {
	case bson.A:
		return e.NormalizeArray(rep)
	case bson.D:
		return e.NormalizeDocument(rep)
	}
	return rep
}

type Observable interface {
	IsEntangled() bool
	LastObserved() time.Time
	InitialObserved() time.Time
}

type Model interface {
	CollapseID() bson.ObjectID
}

type BaseModel struct {
	// KMMeta      ModelMeta     `xml:"-" json:"-" bson:"-" kdb:"-" kcol:"-"`
	ID          bson.ObjectID `xml:"id,attr" json:"ID" bson:"_id,omitempty"`
	UpdatedTime time.Time     `xml:"updated" json:"updated" bson:"updated_time"`
	CreatedTime time.Time     `xml:"created" json:"created" bson:"created_time"`
}

func (e *BaseModel) CollapseID() bson.ObjectID {
	if e.ID.IsZero() {
		e.ID = bson.NewObjectID()
		e.CreatedTime = time.Now()
	}
	e.UpdatedTime = time.Now()
	return e.ID
}

func (e BaseModel) IsEntangled() bool {
	return !e.ID.IsZero()
}

func (e BaseModel) LastObserved() time.Time {
	return e.UpdatedTime
}

func (e BaseModel) InitialObserved() time.Time {
	return e.CreatedTime
}

func GetMetadata(obj Observable) ModelMeta {
	field, found := reflect.TypeOf(obj).FieldByName("KMMeta")
	if !found {
		panic("KMMeta not found")
	}
	return ModelMeta{field.Tag.Get("kdb"), field.Tag.Get("kcol")}
}

func Filter[T Observable](filter model.QueryPredicate) *EntityRecord[T] {
	return All[T]().Filter(filter)
}

func All[T Observable]() *EntityRecord[T] {
	var template T
	recording := &EntityRecord[T]{
		Type: GetMetadata(template),
	}
	return recording
}

func kv(key string, value any) bson.E {
	return bson.E{Key: key, Value: value}
}
