package kosmos

import (
	"time"

	"github.com/borghives/kosmos-go/model"
	"github.com/borghives/kosmos-go/observation"
	"github.com/borghives/kosmos-go/observation/expression"
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

func (e *BaseModel) Collapse() observation.Ripple {
	e.CollapseID()
	e.UpdatedTime = time.Now()
	ripple := observation.Ripple{}
	return *ripple.OnInsertRipple("created_time", e.UpdatedTime)
}

func (e *BaseModel) GetScope() observation.Scope {
	return observation.Scope{} // no witness scope to filter for base model
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

func (e *BaseModel) Decohere(ripple observation.Ripple) {
	if ripple.WasInserted() {
		e.CreatedTime = ripple.GetOnInsertFor("created_time").(time.Time)
	}
}

func Fld(name string) observation.EntityField {
	return observation.EntityField{Name: name}
}

func Filter[T observation.Observable](filters ...expression.QueryFieldPredicate) *observation.EntityDetector[T] {
	return All[T]().Filter(filters...)
}

func All[T observation.Observable]() *observation.EntityDetector[T] {
	var template T
	return observation.NewEntityDetector[T](model.GetMetadata(template))
}

func Witness[C observation.Collapsible](obj C) C {
	observer := observation.NewEntityObserver[C](model.GetMetadata(obj))
	observer.Witness(obj)
	return obj
}
