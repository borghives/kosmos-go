package matter

import (
	"git.mypierian.com/borghives/kosmos-go/meta"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Dataverse struct {
	EntityMeta meta.MetaState
}

func (e *Dataverse) DataCollection() *mongo.Collection {
	return SummonMongo(PurposeAffinityObserver).
		BranchDatabase(e.EntityMeta.BranchName).
		Collection(e.EntityMeta.DataName)
}
