package kosmos

import (
	"context"
	"fmt"
	"log"
	"sync/atomic"
	"time"

	"github.com/borghives/kosmos-go/observation"
	"github.com/borghives/kosmos-go/observation/expression"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

const (
	ModelState_Unset int32 = iota
	ModelState_Transition
	ModelState_Material
)

// Usage Embed BaseModel to your model struct as the first field with kdb and kcol tags
// Example: kosmos.BaseModel `bson:",inline" kdb:"pieriansea" kcol:"page"`
type BaseModel struct {
	ID          bson.ObjectID `xml:"id,attr" json:"ID" bson:"_id,omitempty"`
	UpdatedTime time.Time     `xml:"updated" json:"updated" bson:"updated_time"`
	CreatedTime *time.Time    `xml:"created" json:"created" bson:"created_time,omitempty"`
	state       int32         `xml:"-" json:"-" bson:"-"`
}

func (e *BaseModel) CollapseID() bson.ObjectID {
	if e.ID.IsZero() {
		e.ID = bson.NewObjectID()
	}

	return e.ID
}

func (e *BaseModel) createRipple() observation.Ripple {
	curr := atomic.LoadInt32(&e.state)
	if curr == ModelState_Unset {
		// Unset -> Transition
		if atomic.CompareAndSwapInt32(&e.state, ModelState_Unset, ModelState_Transition) {
			state := observation.RippleState_FromUnknown
			if e.HasID() {
				state = observation.RippleState_FromKnown
			}

			return observation.Ripple{
				State: state,
			}
		}
	}

	if curr == ModelState_Material {
		// Material -> Transition
		if atomic.CompareAndSwapInt32(&e.state, ModelState_Material, ModelState_Transition) {
			return observation.Ripple{
				State: observation.RippleState_FromKnown,
			}
		}
	}

	return observation.Ripple{
		State: observation.RippleState_Unobservable, //A model is unobservable in transition state
	}
}

func (e *BaseModel) Collapse() observation.Ripple {
	ripple := e.createRipple()

	// return early here if model unobservable
	if ripple.State == observation.RippleState_Unobservable {
		return ripple
	}

	ripple.ID = e.ID

	now := time.Now()
	ripple.Set("updated_time", now)
	ripple.Set("created_time", e.CreatedTime) // save old created time during collapse since it is unknown until after decoherence
	e.CreatedTime = nil                       // reset created time just incase of new creation.  Will get back value after decoherence of the ripple.  In short the created time is unknown during transition state.
	return *ripple.OnInsertRipple("created_time", now)
}

func (e *BaseModel) Decohere(ripple observation.Ripple) error {
	//Do Nothing on an unobservable ripple.  We should not be here Decohering an unobservable ripple.
	if ripple.State == observation.RippleState_Unobservable {
		return fmt.Errorf("Failed Decoherence: ripple unobservable.")
	}

	//Materialize the Model
	if atomic.CompareAndSwapInt32(&e.state, ModelState_Transition, ModelState_Material) {
		if ripple.WasInserted() {
			//Was observed and created as new
			createdTime := ripple.GetOnInsertFor("created_time", time.Now()).(time.Time)
			e.CreatedTime = &createdTime // set created time back after decoherence
		} else {
			//Was observed as a previously known entity
			//Set back the created_time saved during Collapse
			createdTime, ok := ripple.Get("created_time")
			if ok {
				e.CreatedTime = createdTime.(*time.Time)
			}
		}

		//Confirm the updated time after a successful state transition
		updatedTime, ok := ripple.Get("updated_time")
		if ok {
			e.UpdatedTime = updatedTime.(time.Time)
		}

		id := ripple.GetID()
		if !id.IsZero() {
			e.ID = id
		}
	}
	return nil
}

func (e BaseModel) SelfScope() expression.Scope {
	return expression.Scope{} // no scope to filter for base model
}

func (e BaseModel) HasID() bool {
	return !e.ID.IsZero()
}

func (e BaseModel) GetID() bson.ObjectID {
	return e.ID
}

func (e BaseModel) LastObserved() time.Time {
	return e.UpdatedTime
}

func Fld(name string) observation.EntityField {
	return observation.EntityField{Name: name}
}

func Filter[T observation.Detectable](filters ...expression.QueryFieldPredicate) *observation.EntityDetector[T] {
	return All[T]().Filter(filters...)
}

func All[T observation.Detectable]() *observation.EntityDetector[T] {
	return observation.NewEntityDetector[T]()
}

func Witness[C observation.Collapsible](ctx context.Context, obj C) error {
	observer := observation.NewEntityObserver[C]()
	return observer.Witness(ctx, obj)
}

func MustHaveObserverClient() *mongo.Client {
	client := observation.SummonMongo(observation.PurposeAffinityObserver).Client()
	if client == nil {
		log.Fatalf("Observer client not initialized")
	}
	return client
}
