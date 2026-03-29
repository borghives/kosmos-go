package ether

import (
	"sync"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	mongoObserverConstants *MongoObserverConstants
	mongoObserverOnce      sync.Once
)

type PurposeAffinity int

const (
	PurposeAffinityUnknown PurposeAffinity = iota
	PurposeAffinityObserver
	PurposeAffinityCreator
	PurposeAffinityAdmin
	PurposeAffinityCount //Max Count/Value for PurposeAffinity
)

func ColapseObserverConstants() *MongoObserverConstants {
	mongoObserverOnce.Do(func() {
		mongoObserverConstants = &MongoObserverConstants{}
		mongoObserverConstants.Coalesce()
	})
	return mongoObserverConstants
}

type MongoObserverConstants struct {
	CmdUri     string // <-- comes from App's Command URI
	Uri        string `mapstructure:"MONGODB_URI"`
	AdminUri   string `mapstructure:"MONGODB_ADMIN_URI"`
	CreatorUri string `mapstructure:"MONGODB_CREATOR_URI"`
}

func (c *MongoObserverConstants) Coalesce() Ether {
	viper.BindEnv("MONGODB_URI")
	viper.BindEnv("MONGODB_ADMIN_URI")
	viper.BindEnv("MONGODB_CREATOR_URI")
	viper.AutomaticEnv()
	viper.Unmarshal(&c)

	return c
}

func (c *MongoObserverConstants) MergeFromFile(filename string) Ether {
	// 1. look for addition environment file
	viper.SetConfigFile(filename)
	_ = viper.MergeInConfig()
	viper.Unmarshal(c)
	return c
}

func (c *MongoObserverConstants) MergeFromCmd(cmd *cobra.Command) Ether {
	return c
}

func (c *MongoObserverConstants) GetURI(purpose PurposeAffinity) string {
	switch purpose {
	case PurposeAffinityObserver:
		return c.Uri
	case PurposeAffinityCreator:
		return c.CreatorUri
	case PurposeAffinityAdmin:
		return c.AdminUri
	default:
		return c.Uri
	}
}
