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

func (r *Ripple) GetOnInsertFor(key string) any {
	for _, expr := range r.Expr {
		if expr.Key == "$setOnInsert" {
			for _, setOnInsertExpr := range expr.Value.(bson.D) {
				if setOnInsertExpr.Key == key {
					return setOnInsertExpr.Value
				}
			}
		}
	}
	return nil
}

func (r *Ripple) WasInserted() bool {
	return r.InsertFeedback != nil || (r.UpdateFeedback != nil && r.UpdateFeedback.UpsertedID != nil)
}

type Collapsible interface {
	IsEntangled() bool
	GetScope() Scope //return the scope of the Collapse
	CollapseID() bson.ObjectID
	Collapse() Ripple       //return the ripple side effect after the collapse.  This will implicitly collapse the ID
	Decohere(ripple Ripple) //After the collapse and interaction with environment, an entity decoheres (ripple contains materialization info)
}
