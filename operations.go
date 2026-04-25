package kosmos

import (
	"context"
	"log"

	"github.com/borghives/kosmos-go/observation"
	"github.com/borghives/kosmos-go/observation/expression"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func Fld(name string) observation.EntityField {
	return observation.EntityField{Name: name}
}

func All[T observation.Detectable]() *observation.EntityDetector[T] {
	return observation.NewEntityDetector[T]()
}

func Detect[T observation.Detectable](filters ...expression.QueryFieldPredicate) *observation.EntityDetector[T] {
	return All[T]().Filter(filters...)
}

func Witness[C observation.Collapsible](ctx context.Context, obj C) error {
	observer := observation.NewEntityObserver[C]()
	return observer.Witness(ctx, obj)
}

func MustHaveObserverClient() *mongo.Client {
	client := observation.SummonMongo(observation.PurposeAffinityObserver).Client()
	if client == nil {
		log.Fatalf("Observer client not initialized")
	}
	return client
}
