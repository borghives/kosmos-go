package expression

import (
	"github.com/borghives/kosmos-go/model"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Scope []QueryFieldPredicate

func CreateScope(filters ...QueryFieldPredicate) Scope {
	return Scope(filters)
}

func (s Scope) Express(objMeta model.Metadata) bson.D {
	if len(s) == 0 {
		return bson.D{}
	} else if len(s) == 1 {
		return NormalizeExpression(s[0], objMeta.ResolveAlias).(bson.D)
	} else {
		exprs := ToBSONArray(s...)
		return NormalizeExpression(And(exprs), objMeta.ResolveAlias).(bson.D)
	}
}
