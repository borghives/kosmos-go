package observation

import (
	"context"
	"fmt"

	"github.com/borghives/kosmos-go/model"
	"github.com/borghives/kosmos-go/observation/expression"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type EntityQuery[T model.Observable] struct {
	Type   model.Metadata
	stages Aggregation
}

func (r *EntityQuery[T]) Filter(filters ...expression.QueryFieldPredicate) *EntityQuery[T] {
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

func (r *EntityQuery[T]) FilterEither(filters ...expression.QueryFieldPredicate) *EntityQuery[T] {
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

func (r *EntityQuery[T]) Sort(field string, descending bool) *EntityQuery[T] {
	order := 1
	if descending {
		order = -1
	}
	r.stages = r.stages.Sort(bson.D{kv(field, order)})
	return r
}

func (r *EntityQuery[T]) PullOne() *T {
	results, err := r.pullPipeline(Aggregation{}.Limit(1))
	if err != nil {
		return nil
	}
	if len(results) == 0 {
		return nil
	}
	return &results[0]
}

func (r *EntityQuery[T]) PullAll() []T {
	results, err := r.pullPipeline(Aggregation{})
	if err != nil {
		return nil
	}
	return results
}

func (r *EntityQuery[T]) dataCollection() *mongo.Collection {
	observer := SummonMongo(PurposeAffinityObserver)
	return observer.Database(r.Type.DatabaseName).Collection(r.Type.CollectionName)
}

func (r *EntityQuery[T]) PipelineJSON() string {
	return r.stages.JsonString()
}

func (r *EntityQuery[T]) pullPipeline(postStages Aggregation) ([]T, error) {
	collection := r.dataCollection()
	pipeline := r.stages.AppendFrom(postStages).Pipeline()
	fmt.Println(r.stages.AppendFrom(postStages).JsonString())
	cursor, err := collection.Aggregate(context.Background(), pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var results []T
	cursor.All(context.Background(), &results)
	return results, nil
}
