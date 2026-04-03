package observation

import (
	"context"
	"fmt"
	"time"

	"github.com/borghives/kosmos-go/model"
	"github.com/borghives/kosmos-go/observation/expression"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Observable interface {
	IsEntangled() bool
	LastObserved() time.Time
	InitialObserved() time.Time
}

type EntityDetector[T Observable] struct {
	EntityDataverse
	stages Aggregation
}

func NewEntityDetector[T Observable](entityMeta model.Metadata) *EntityDetector[T] {
	return &EntityDetector[T]{
		EntityDataverse: EntityDataverse{EntityMeta: entityMeta},
	}
}

func (r *EntityDetector[T]) Filter(filters ...expression.QueryFieldPredicate) *EntityDetector[T] {
	if len(filters) == 0 {
		return r
	} else if len(filters) == 1 {
		r.stages = r.stages.Match(expression.NormalizeExpression(filters[0], r.EntityMeta.ResolveAlias).(bson.D))
	} else {
		exprs := make(bson.A, len(filters))
		for i, f := range filters {
			exprs[i] = f
		}
		r.stages = r.stages.Match(expression.NormalizeExpression(expression.And(exprs), r.EntityMeta.ResolveAlias).(bson.D))
	}
	return r
}

func (r *EntityDetector[T]) FilterEither(filters ...expression.QueryFieldPredicate) *EntityDetector[T] {
	if len(filters) == 0 {
		return r
	} else if len(filters) == 1 {
		r.stages = r.stages.Match(expression.NormalizeExpression(filters[0], r.EntityMeta.ResolveAlias).(bson.D))
	} else {
		exprs := make(bson.A, len(filters))
		for i, f := range filters {
			exprs[i] = f
		}
		r.stages = r.stages.Match(expression.NormalizeExpression(expression.Or(exprs), r.EntityMeta.ResolveAlias).(bson.D))
	}
	return r
}

func (r *EntityDetector[T]) Sort(field string, descending bool) *EntityDetector[T] {
	order := 1
	if descending {
		order = -1
	}
	r.stages = r.stages.Sort(bson.D{kv(field, order)})
	return r
}

func (r *EntityDetector[T]) PullOne() (*T, error) {
	results, err := r.pullPipeline(context.Background(), Aggregation{}.Limit(1))
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, nil
	}

	return &results[0], nil
}

func (r *EntityDetector[T]) PullAll() ([]T, error) {
	results, err := r.pullPipeline(context.Background(), Aggregation{})
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (r *EntityDetector[T]) PipelineJSON() string {
	return r.stages.JsonString()
}

func (r *EntityDetector[T]) RunPipeline(ctx context.Context, postStages Aggregation) (*mongo.Cursor, error) {
	dataCollection := r.DataCollection()

	stages := r.stages.AppendFrom(postStages)
	cursor, err := dataCollection.Aggregate(ctx, stages.Pipeline())
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate %v: %v", stages.JsonString(), err)
	}
	return cursor, nil
}

func (r *EntityDetector[T]) pullPipeline(ctx context.Context, postStages Aggregation) ([]T, error) {
	cursor, err := r.RunPipeline(ctx, postStages)
	if err != nil {
		return nil, fmt.Errorf("failed to pull: %v", err)
	}

	var results []T
	err = cursor.All(ctx, &results)
	if err != nil {
		return nil, fmt.Errorf("failed to decode results: %v", err)
	}

	return results, nil
}
