package observation

import (
	"context"

	"github.com/borghives/kosmos-go/model"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func OnInsertRipple(key string, value any) model.Ripple {
	return model.Ripple{kv("$setOnInsert", bson.D{kv(key, value)})}
}

type EntityObserver[T model.Collapsable] struct {
	Type model.Metadata
}

func (r *EntityObserver[T]) dataCollection() *mongo.Collection {
	observer := SummonMongo(PurposeAffinityObserver)
	return observer.Database(r.Type.DatabaseName).Collection(r.Type.CollectionName)
}

func (r *EntityObserver[T]) Witness(object T) {
	scope := object.WitnessScope()
	isEntangled := object.IsEntangled()
	ripple := object.Collapse()

	// if no impact scope to filter by and not entangled, it's a new record
	if len(scope) == 0 && !isEntangled {
		r.dataCollection().InsertOne(context.Background(), object)
		return
	}

	// if entangled, use the collapse id as scope
	if isEntangled {
		scope = model.Scope{kv("_id", object.CollapseID())}
	}

	update := bson.D{kv("$set", object)}
	update = append(update, ripple...) // add ripple affect to update

	updateOption := options.UpdateOne().SetUpsert(true)
	r.dataCollection().UpdateOne(context.Background(), scope, update, updateOption)
}
