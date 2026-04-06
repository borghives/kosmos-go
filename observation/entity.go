package observation

import (
	"github.com/borghives/kosmos-go/model"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type EntityDataverse struct {
	EntityMeta model.Metadata
}

func (e *EntityDataverse) DataCollection() *mongo.Collection {
	return SummonMongo(PurposeAffinityObserver).
		BranchDatabase(e.EntityMeta.BranchName).
		Collection(e.EntityMeta.DataName)
}

func (e *EntityDataverse) PingClient() {
	SummonMongo(PurposeAffinityObserver).Client()
}
