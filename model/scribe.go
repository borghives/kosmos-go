package model

import (
	"context"
	"time"

	"github.com/borghives/kosmos-go/observation"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Observable interface {
	IsEntangled() bool
	LastObserved() time.Time
	InitialObserved() time.Time
}

type Scope bson.D
type Ripple bson.D

func OnInsertRipple(key string, value any) Ripple {
	return Ripple{kv("$setOnInsert", bson.D{kv(key, value)})}
}

type Collapsable interface {
	IsEntangled() bool
	CollapseID() bson.ObjectID
	Collapse() Ripple   //return the ripple side effect after the collapse.  This will implicitly collapse the ID
	WitnessScope() Scope //return the scope to filter by
}

type EntityObserver[T Collapsable] struct {
	Type Metadata
}

func (r *EntityObserver[T]) dataCollection() *mongo.Collection {
	observer := observation.SummonMongo(observation.PurposeAffinityObserver)
	return observer.Database(r.Type.DatabaseName).Collection(r.Type.CollectionName)
}

func (r *EntityObserver[T]) Witness(model T) {
	scope := model.WitnessScope()
	isEntangled := model.IsEntangled()
	ripple := model.Collapse()

	// if no impact scope to filter by and not entangled, it's a new record
	if len(scope) == 0 && !isEntangled {
		r.dataCollection().InsertOne(context.Background(), model)
		return
	}

	// if entangled, use the collapse id as scope
	if isEntangled {
		scope = Scope{kv("_id", model.CollapseID())}
	}

	update := bson.D{kv("$set", model)}
	update = append(update, ripple...) // add ripple affect to update

	updateOption := options.UpdateOne().SetUpsert(true)
	r.dataCollection().UpdateOne(context.Background(), scope, update, updateOption)
}
