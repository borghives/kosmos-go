package observation

import (
	"encoding/json"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Aggregation struct {
	pipeline mongo.Pipeline
}

func (a Aggregation) Match(filter bson.D) Aggregation {
	return Aggregation{pipeline: append(a.pipeline, bson.D{kv("$match", filter)})}
}

func (a Aggregation) Group(group bson.D) Aggregation {
	return Aggregation{pipeline: append(a.pipeline, bson.D{kv("$group", group)})}
}

func (a Aggregation) Lookup(lookup bson.D) Aggregation {
	return Aggregation{pipeline: append(a.pipeline, bson.D{kv("$lookup", lookup)})}
}

func (a Aggregation) AddFields(field bson.D) Aggregation {
	return Aggregation{pipeline: append(a.pipeline, bson.D{kv("$addFields", field)})}
}

func (a Aggregation) Project(fields bson.D) Aggregation {
	return Aggregation{pipeline: append(a.pipeline, bson.D{kv("$project", fields)})}
}

func (a Aggregation) Sort(fields bson.D) Aggregation {
	return Aggregation{pipeline: append(a.pipeline, bson.D{kv("$sort", fields)})}
}

func (a Aggregation) Limit(value int64) Aggregation {
	return Aggregation{pipeline: append(a.pipeline, bson.D{kv("$limit", value)})}
}

func (a Aggregation) Search(fields bson.D) Aggregation {
	return Aggregation{pipeline: append(a.pipeline, bson.D{kv("$search", fields)})}
}

func (a Aggregation) AppendFrom(agg Aggregation) Aggregation {
	return Aggregation{pipeline: append(a.pipeline, agg.pipeline...)}
}

func (a Aggregation) Pipeline() mongo.Pipeline {
	return a.pipeline
}

// mainly for debugging
func (a Aggregation) JsonString() string {
	// Marshal bson.A to JSON
	jsonString, err := json.Marshal(a.pipeline)
	if err != nil {
		return fmt.Sprintf("Error marshaling aggregation pipeline: %v", err)
	}

	return string(jsonString)
}
