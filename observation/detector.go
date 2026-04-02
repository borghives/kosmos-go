package observation

import (
	"context"
	"fmt"

	"github.com/borghives/kosmos-go/model"
	"github.com/borghives/kosmos-go/observation/expression"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type EntityDetector[T model.Observable] struct {
	Type   model.Metadata
	stages Aggregation
}

func (r *EntityDetector[T]) Filter(filters ...expression.QueryFieldPredicate) *EntityDetector[T] {
	if len(filters) == 0 {
		return r
	} else if len(filters) == 1 {
		r.stages = r.stages.Match(expression.NormalizeExpression(filters[0], r.Type.ResolveAlias).(bson.D))
	} else {
		exprs := make(bson.A, len(filters))
		for i, f := range filters {
			exprs[i] = f
		}
		r.stages = r.stages.Match(expression.NormalizeExpression(expression.And(exprs), r.Type.ResolveAlias).(bson.D))
	}
	return r
}

func (r *EntityDetector[T]) FilterEither(filters ...expression.QueryFieldPredicate) *EntityDetector[T] {
	if len(filters) == 0 {
		return r
	} else if len(filters) == 1 {
		r.stages = r.stages.Match(expression.NormalizeExpression(filters[0], r.Type.ResolveAlias).(bson.D))
	} else {
		exprs := make(bson.A, len(filters))
		for i, f := range filters {
			exprs[i] = f
		}
		r.stages = r.stages.Match(expression.NormalizeExpression(expression.Or(exprs), r.Type.ResolveAlias).(bson.D))
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

func (r *EntityDetector[T]) dataCollection() *mongo.Collection {
	return SummonMongo(PurposeAffinityObserver).Collection(r.Type.CollectionName)
}

func (r *EntityDetector[T]) PipelineJSON() string {
	return r.stages.JsonString()
}

func (r *EntityDetector[T]) pullPipeline(ctx context.Context, postStages Aggregation) ([]T, error) {
	dataCollection := r.dataCollection()

	stages := r.stages.AppendFrom(postStages)
	cursor, err := dataCollection.Aggregate(ctx, stages.Pipeline())
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate %v: %v", stages.JsonString(), err)
	}
	defer cursor.Close(ctx)

	var results []T
	err = cursor.All(ctx, &results)
	if err != nil {
		return nil, fmt.Errorf("failed to decode results: %v", err)
	}

	return results, nil
}
