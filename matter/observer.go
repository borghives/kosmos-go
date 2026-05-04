package matter

import (
	"context"
	"fmt"

	"github.com/borghives/kosmos-go/meta"
	"github.com/borghives/kosmos-go/meta/expression"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// An Observer record information

type Observer[T Collapsible] struct {
	Dataverse
}

func NewObserver[T Collapsible]() *Observer[T] {
	var template T
	return &Observer[T]{
		Dataverse: Dataverse{EntityMeta: meta.GetMetadata(template)},
	}
}

func (r *Observer[T]) Record(ctx context.Context, object T) error {

	ripple := object.Collapse() //after a collapse the object will be in a transitional state and the ID is materialized
	if ripple.State == RippleState_Unobservable {
		return fmt.Errorf("The model could not be observed or collapse")
	}

	//NOTE!!: Early exit without the object Decohere[nce] at the end will keep the object in a transitional (INBETWEEN) state

	// if no Self filter and from an unknown state, insert as new record
	scope := object.SelfScope()
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
			id := ripple.ID
			if id.IsZero() {
				id = object.CollapseID()
			}
			scope = expression.CreateScope(EntityField{"_id"}.Eq(id))
		}

		scopeExpr := scope.Express(r.EntityMeta)
		update := bson.D{kv("$set", object)}
		update = append(update, ripple.Expr...) // add ripple affect to update

		updateOption := options.UpdateOne().SetUpsert(true)
		updateResult, err := r.DataCollection().UpdateOne(ctx, scopeExpr, update, updateOption)
		if err != nil {
			return err
		}
		ripple.UpdateFeedback = updateResult
	}

	return object.Decohere(ripple)
}
