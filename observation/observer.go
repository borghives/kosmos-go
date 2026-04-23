package observation

import (
	"context"
	"fmt"

	"github.com/borghives/kosmos-go/model"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// An Observer witness/memorize information

type EntityObserver[T Collapsible] struct {
	EntityDataverse
}

func NewEntityObserver[T Collapsible]() *EntityObserver[T] {
	var template T
	return &EntityObserver[T]{
		EntityDataverse: EntityDataverse{EntityMeta: model.GetMetadata(template)},
	}
}

func (r *EntityObserver[T]) Witness(ctx context.Context, object T) error {

	ripple := object.Collapse() //after a collapse the object will be in a transitional state and the ID is materialized
	if ripple.State == RippleState_Unobservable {
		return fmt.Errorf("The model could not be observed or collapse")
	}

	//NOTE!!: Early exit without reaching the object Decohere[nce] at the end will keep the object in a transitional state
	scope := object.SelfScope()

	// if no Self filter and from an unknown state, insert as new record
	if len(scope) == 0 && ripple.State == RippleState_FromUnknown {
		insertResult, err := r.DataCollection().InsertOne(ctx, object)
		if err != nil {
			return err
		}
		ripple.InsertFeedback = insertResult
	} else {
		// using scope to find existing record
		// if self scope is empty, default to using ID
		if len(scope) == 0 {
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

	return object.Decohere(ripple)
}
