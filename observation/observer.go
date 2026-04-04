package observation

import (
	"context"

	"github.com/borghives/kosmos-go/model"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type EntityObserver[T Collapsible] struct {
	EntityDataverse
}

func NewEntityObserver[T Collapsible](entityMeta model.Metadata) *EntityObserver[T] {
	return &EntityObserver[T]{
		EntityDataverse: EntityDataverse{EntityMeta: entityMeta},
	}
}

func (r *EntityObserver[T]) Witness(ctx context.Context, object T) error {
	scope := object.GetScope()
	isEntangled := object.IsEntangled()
	ripple := object.Collapse()

	// if no scope to filter by and not previously entangled, it's a new record
	if len(scope) == 0 && !isEntangled {
		insertResult, err := r.DataCollection().InsertOne(ctx, object)
		if err != nil {
			return err
		}
		ripple.InsertFeedback = insertResult
	} else {
		// if entangled, use the collapse id as scope
		if isEntangled {
			scope = Scope{kv("_id", object.CollapseID())}
		}

		update := bson.D{kv("$set", object)}
		update = append(update, ripple.Expr...) // add ripple affect to update

		updateOption := options.UpdateOne().SetUpsert(true)
		updateResult, err := r.DataCollection().UpdateOne(ctx, scope, update, updateOption)
		if err != nil {
			return err
		}
		ripple.UpdateFeedback = updateResult
	}

	object.Decohere(ripple)
	return nil
}
