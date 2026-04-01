package kosmos

import (
	"time"

	"github.com/borghives/kosmos-go/model"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// Usage Embed BaseModel to your model struct as the first field with kdb and kcol tags
// Example: kosmos.BaseModel `bson:",inline" kdb:"pieriansea" kcol:"page"`
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

	return e.ID
}

func (e *BaseModel) Collapse() model.Ripple {
	e.CollapseID()
	e.UpdatedTime = time.Now()
	ripple := model.OnInsertRipple("created_time", e.CreatedTime)
	return ripple
}

func (e *BaseModel) WitnessScope() model.Scope {
	return model.Scope{} // no witness scope to filter for base model
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

func Filter[T model.Observable](filter model.QueryPredicate) *model.EntityRecord[T] {
	return All[T]().Filter(filter)
}

func All[T model.Observable]() *model.EntityRecord[T] {
	var template T
	recording := &model.EntityRecord[T]{
		Type: model.GetMetadata(template),
	}
	return recording
}

func Witness[C model.Collapsable](obj C) C {
	observer := &model.EntityObserver[C]{
		Type: model.GetMetadata(obj),
	}
	observer.Witness(obj)
	return obj
}
