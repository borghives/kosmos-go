package observation

import (
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Scope bson.D
type Ripple struct {
	Expr           bson.D
	InsertFeedback *mongo.InsertOneResult
	UpdateFeedback *mongo.UpdateResult
}

func (r *Ripple) OnInsertRipple(key string, value any) *Ripple {
	if r.Expr == nil {
		r.Expr = bson.D{}
	}
	r.Expr = append(r.Expr, kv("$setOnInsert", bson.D{kv(key, value)}))
	return r
}

type Collapsible interface {
	IsEntangled() bool
	GetCollapseScope() Scope //return the scope of the Collapse
	CollapseID() bson.ObjectID
	Collapse() Ripple       //return the ripple side effect after the collapse.  This will implicitly collapse the ID
	Decohere(ripple Ripple) //After the collapse and interaction with environment, an entity decoheres (ripple contains materialization info)
}
