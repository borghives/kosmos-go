package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Entity struct {
	ID          bson.ObjectID `xml:"id,attr" json:"ID" bson:"_id,omitempty"`
	UpdatedTime time.Time     `xml:"updated" json:"updated" bson:"updated_time"`
}
