package ether

import (
	"sync"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	mongoConstants *MongoConstants
	mongoOnce      sync.Once
)

func CollapseDataverseConstants() *MongoConstants {
	mongoOnce.Do(func() {
		CollapseConstants()
		mongoConstants = &MongoConstants{}
		mongoConstants.Coalesce()
	})
	return mongoConstants
}

type MongoConstants struct {
	CmdUri     string // <-- comes from App's Command URI
	Uri        string `mapstructure:"MONGODB_URI"`
	AdminUri   string `mapstructure:"MONGODB_ADMIN_URI"`
	CreatorUri string `mapstructure:"MONGODB_CREATOR_URI"`
	Database   string `mapstructure:"MONGODB_DATABASE"`
}

func (c *MongoConstants) Coalesce() Ether {
	viper.BindEnv("MONGODB_URI")
	viper.BindEnv("MONGODB_ADMIN_URI")
	viper.BindEnv("MONGODB_CREATOR_URI")
	viper.BindEnv("MONGODB_DATABASE")
	viper.AutomaticEnv()
	viper.Unmarshal(&c)

	return c
}

func (c *MongoConstants) MergeFromFile(filename string) Ether {
	// 1. look for addition environment file
	viper.SetConfigFile(filename)
	_ = viper.MergeInConfig()
	viper.Unmarshal(c)
	return c
}

func (c *MongoConstants) MergeFromCmd(cmd *cobra.Command) Ether {
	if flag := cmd.Flags().Lookup("uri"); flag != nil {
		viper.BindPFlag("MONGODB_URI", flag)
		c.CmdUri = flag.Value.String()
	}
	viper.Unmarshal(c)
	return c
}
