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

func (e *ModelMeta) NormalizeArray(array bson.A) bson.A {
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

// Usage Embed BaseModel to your model struct as the first field with kdb and kcol tags
// Example: kosmos.BaseModel `xml:"-" json:"-" bson:"inline" kdb:"pieriansea" kcol:"page"`
type BaseModel struct {
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
	field, found := reflect.TypeOf(obj).FieldByName("BaseModel")
	if !found {
		panic("BaseModel not found")
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
