package observation

import (
	"github.com/borghives/kosmos-go/meta"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type EntityDataverse struct {
	EntityMeta meta.Metadata
}

func (e *EntityDataverse) DataCollection() *mongo.Collection {
	return SummonMongo(PurposeAffinityObserver).
		BranchDatabase(e.EntityMeta.BranchName).
		Collection(e.EntityMeta.DataName)
}
