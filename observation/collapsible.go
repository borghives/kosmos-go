package observation

import (
	"github.com/borghives/kosmos-go/observation/expression"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type RippleState int32

const (
	RippleState_Unobservable RippleState = iota //failed state
	RippleState_FromUnknown                     //Forming into existence
	RippleState_FromKnown                       //Previously observed
)

type Ripple struct {
	ID             bson.ObjectID
	State          RippleState
	Expr           bson.D
	Interstitial   map[string]any //state during the period of transition.
	InsertFeedback *mongo.InsertOneResult
	UpdateFeedback *mongo.UpdateResult
}

type Collapsible interface {
	CollapseID() bson.ObjectID
	HasID() bool
	SelfScope() expression.Scope //return the scope of Self.  The scope should have enough information to UNIQUELY filter itself

	//An entity must be Collapse by an observer and Decohere the interaction ripple in order to exists in a known state.
	//A failure of Collapse AND Decohere flow will put the state of the object in an UNKNOWN / INBETWEEN state.
	Collapse() Ripple             //return the ripple side effect after the collapse. Once Collapse is called the model is in an INBETWEEN state
	Decohere(ripple Ripple) error //After the collapse and interaction with environment, an entity decoheres (ripple contains materialization info)
}

func (r *Ripple) Set(key string, value any) *Ripple {
	if r.Interstitial == nil {
		r.Interstitial = make(map[string]any)
	}
	r.Interstitial[key] = value
	return r
}

func (r *Ripple) Get(key string) (any, bool) {
	value, ok := r.Interstitial[key]
	return value, ok
}

func (r *Ripple) OnInsertRipple(key string, value any) *Ripple {
	if r.Expr == nil {
		r.Expr = bson.D{}
	}
	r.Expr = append(r.Expr, kv("$setOnInsert", bson.D{kv(key, value)}))
	return r
}

func (r *Ripple) GetOnInsertFor(key string, defaultValue any) any {
	for _, expr := range r.Expr {
		if expr.Key == "$setOnInsert" {
			for _, setOnInsertExpr := range expr.Value.(bson.D) {
				if setOnInsertExpr.Key == key {
					return setOnInsertExpr.Value
				}
			}
		}
	}
	return defaultValue
}

func (r Ripple) WasInserted() bool {
	return r.InsertFeedback != nil || (r.UpdateFeedback != nil && r.UpdateFeedback.UpsertedID != nil)
}

func (r Ripple) GetID() bson.ObjectID {
	if r.InsertFeedback != nil {
		return r.InsertFeedback.InsertedID.(bson.ObjectID)
	}

	if r.UpdateFeedback != nil && r.UpdateFeedback.UpsertedID != nil {
		return r.UpdateFeedback.UpsertedID.(bson.ObjectID)
	}

	return bson.NilObjectID
}
