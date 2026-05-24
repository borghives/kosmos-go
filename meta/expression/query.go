package expression

import "go.mongodb.org/mongo-driver/v2/bson"

type QueryFieldPredicate struct {
	FieldName FieldName
	Query     QueryOp
}

func (q QueryFieldPredicate) Empty() bool {
	return q.FieldName.Name == "" && q.Query.Operator == ""
}

func (q QueryFieldPredicate) ToRepr() any {
	if q.Empty() {
		return bson.D{}
	}
	return bson.D{kv(q.FieldName.Name, q.Query.ToRepr())}
}

func (q QueryFieldPredicate) Reduce(resolver NameResolver) any {
	if q.Empty() {
		return bson.D{}
	}
	return bson.D{kv(resolver(q.FieldName.Name), NormalizeExpression(q.Query, resolver))}
}

// ***** QUERY OPs*******

type QueryOp struct {
	Operator string
	Value    any
}

func (q QueryOp) ToRepr() any {
	return bson.D{kv(q.Operator, q.Value)}
}

func Eq(value any) QueryOp {
	return QueryOp{"$eq", value}
}

func Ne(value any) QueryOp {
	return QueryOp{"$ne", value}
}

func Gt(value any) QueryOp {
	return QueryOp{"$gt", value}
}

func Gte(value any) QueryOp {
	return QueryOp{"$gte", value}
}

func Lt(value any) QueryOp {
	return QueryOp{"$lt", value}
}

func Lte(value any) QueryOp {
	return QueryOp{"$lte", value}
}

func In(values bson.A) QueryOp {
	return QueryOp{"$in", values}
}

func Nin(values bson.A) QueryOp {
	return QueryOp{"$nin", values}
}

func And(values bson.A) QueryOp {
	return QueryOp{"$and", values}
}

func Or(values bson.A) QueryOp {
	return QueryOp{"$or", values}
}

// ***** Operator *******

func ToOpParamExpression(input any) any {
	switch input := input.(type) {
	case int:
		return LiteralValue{Value: input}
	case string:
		return FieldPath{Path: input}
	case Base:
		return input
	}
	return input
}

type Op struct {
	Operator string
	Param    any
}

func (q Op) ToRepr() any {
	return bson.D{kv(q.Operator, q.Param)}
}

func Sum(input any) Op {
	return Op{
		Operator: "$sum",
		Param:    ToOpParamExpression(input),
	}
}

func Avg(input any) Op {
	return Op{
		Operator: "$avg",
		Param:    ToOpParamExpression(input),
	}
}

func Min(input any) Op {
	return Op{
		Operator: "$min",
		Param:    ToOpParamExpression(input),
	}
}

func Max(input any) Op {
	return Op{
		Operator: "$max",
		Param:    ToOpParamExpression(input),
	}
}

// ***** OR QUERY *******

type OrQueryFieldPredicate struct {
	Queries []QueryFieldPredicate
}

func OrField(query ...QueryFieldPredicate) OrQueryFieldPredicate {
	return OrQueryFieldPredicate{
		Queries: query,
	}
}

func (oq OrQueryFieldPredicate) ToRepr() any {
	switch len(oq.Queries) {
	case 0:
		return bson.D{}
	case 1:
		return oq.Queries[0]
	default:
		return Or(ToBSONArray(oq.Queries...))
	}
}

// ***** And QUERY *******

type AndQueryFieldPredicate struct {
	Queries []QueryFieldPredicate
}

func AndField(query ...QueryFieldPredicate) AndQueryFieldPredicate {
	return AndQueryFieldPredicate{
		Queries: query,
	}
}

func (oq AndQueryFieldPredicate) ToRepr() any {
	switch len(oq.Queries) {
	case 0:
		return bson.D{}
	case 1:
		return oq.Queries[0]
	default:
		return And(ToBSONArray(oq.Queries...))
	}
}

func ToBSONArray(filters ...QueryFieldPredicate) bson.A {
	exprs := make(bson.A, 0, len(filters))
	for _, f := range filters {
		if !f.Empty() {
			exprs = append(exprs, f)
		}
	}
	return exprs
}
