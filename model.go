package kosmos

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/borghives/kosmos-go/matter"
	"github.com/borghives/kosmos-go/meta/expression"
	"go.mongodb.org/mongo-driver/v2/bson"
)

const (
	ModelState_Unset int32 = iota
	ModelState_Transition
	ModelState_Material
)

// Usage Embed BaseModel to your model struct as the first field with a kosmos tag
// Example: kosmos.BaseModel `bson:",inline" kosmos:"pieriansea>page"`
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

func (e *BaseModel) createRipple() matter.Ripple {
	curr := atomic.LoadInt32(&e.state)
	if curr == ModelState_Unset {
		// Unset -> Transition
		if atomic.CompareAndSwapInt32(&e.state, ModelState_Unset, ModelState_Transition) {
			state := matter.RippleState_FromUnknown
			if e.HasID() {
				state = matter.RippleState_FromKnown
			}

			return matter.Ripple{
				State: state,
			}
		}
	}

	if curr == ModelState_Material {
		// Material -> Transition
		if atomic.CompareAndSwapInt32(&e.state, ModelState_Material, ModelState_Transition) {
			return matter.Ripple{
				State: matter.RippleState_FromKnown,
			}
		}
	}

	return matter.Ripple{
		State: matter.RippleState_Unobservable, //A model is unobservable in transition state
	}
}

func (e *BaseModel) Collapse() matter.Ripple {
	ripple := e.createRipple()

	// return early here if model unobservable
	if ripple.State == matter.RippleState_Unobservable {
		return ripple
	}

	ripple.ID = e.ID

	now := time.Now()
	ripple.Set("updated_time", now)
	ripple.Set("created_time", e.CreatedTime) // save old created time during collapse since it is unknown until after decoherence
	e.CreatedTime = nil                       // reset created time just incase of new creation.  Will get back value after decoherence of the ripple.  In short the created time is unknown during transition state.
	e.UpdatedTime = now
	return *ripple.OnInsertRipple("created_time", now)
}

func (e *BaseModel) Decohere(ripple matter.Ripple) error {
	//Do Nothing on an unobservable ripple.  We should not be here Decohering an unobservable ripple.
	if ripple.State == matter.RippleState_Unobservable {
		return fmt.Errorf("Failed Decoherence: ripple state is unobservable.")
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
