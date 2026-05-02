package matter

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/borghives/kosmos-go/klog"
	"github.com/borghives/kosmos-go/matter/expression"
	"github.com/borghives/kosmos-go/meta"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Detectable interface {
	GetID() bson.ObjectID
	HasID() bool
	LastObserved() time.Time
}

type Detector[T Detectable] struct {
	Dataverse
	stages Aggregation
}

func NewDetector[T Detectable]() *Detector[T] {
	var template T
	return &Detector[T]{
		Dataverse: Dataverse{EntityMeta: meta.GetMetadata(template)},
	}
}

func (r *Detector[T]) Filter(filters ...expression.QueryFieldPredicate) *Detector[T] {
	if len(filters) == 0 {
		return r
	} else if len(filters) == 1 {
		r.stages = r.stages.Match(expression.NormalizeExpression(filters[0], r.EntityMeta.ResolveAlias).(bson.D))
	} else {
		exprs := expression.ToBSONArray(filters...)
		r.stages = r.stages.Match(expression.NormalizeExpression(expression.And(exprs), r.EntityMeta.ResolveAlias).(bson.D))
	}
	return r
}

func (r *Detector[T]) FilterEither(filters ...expression.QueryFieldPredicate) *Detector[T] {
	if len(filters) < 2 {
		return r.Filter(filters...)
	}

	r.stages = r.stages.Match(expression.NormalizeExpression(expression.OrField(filters...), r.EntityMeta.ResolveAlias).(bson.D))
	return r
}

func (r *Detector[T]) Limit(limit int64) *Detector[T] {
	r.stages = r.stages.Limit(limit)
	return r
}

func (r *Detector[T]) SortLatest() *Detector[T] {
	return r.Sort("updated_time", true)
}

func (r *Detector[T]) Sort(field string, descending bool) *Detector[T] {
	field = r.EntityMeta.ResolveAlias(field)
	order := 1
	if descending {
		order = -1
	}
	r.stages = r.stages.Sort(bson.D{kv(field, order)})
	return r
}

func (r *Detector[T]) PullOne(ctx context.Context) (*T, error) {
	results, err := r.pullPipeline(ctx, Aggregation{}.Limit(1))
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, nil
	}

	return &results[0], nil
}

func (r Detector[T]) PullAll(ctx context.Context) ([]T, error) {
	results, err := r.pullPipeline(ctx, Aggregation{})
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (r *Detector[T]) PipelineJSON() string {
	return r.stages.JsonString()
}

func (r Detector[T]) RunPipeline(ctx context.Context, postStages Aggregation) (*mongo.Cursor, error) {
	dataCollection := r.DataCollection()

	stages := r.stages.AppendFrom(postStages)
	cursor, err := dataCollection.Aggregate(ctx, stages.Pipeline())
	if err != nil {
		slog.Debug("Failed to aggregate pipe for run", stages.Log(), klog.Err(err))
		return nil, fmt.Errorf("failed to aggregate %v", err)
	}
	return cursor, nil
}

func (r Detector[T]) pullPipeline(ctx context.Context, postStages Aggregation) ([]T, error) {
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
