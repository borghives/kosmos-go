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

func ColapseObserverConstants() *MongoObserverConstants {
	mongoObserverOnce.Do(func() {
		CollapseConstants()
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
	if flag := cmd.Flags().Lookup("uri"); flag != nil {
		viper.BindPFlag("MONGODB_URI", flag)
		c.CmdUri = flag.Value.String()
	}
	viper.Unmarshal(c)
	return c
}
