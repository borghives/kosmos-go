package ether

import (
	"os"
	"sync"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	constants    *Constants
	collapseOnce sync.Once
)

func CollapseConstants() *Constants {
	collapseOnce.Do(func() {
		constants = &Constants{}
		constants.Coalesce()
	})
	return constants
}

type Ether interface {
	Coalesce() Ether
	MergeFromFile(filename string) Ether
	MergeFromCmd(cmd *cobra.Command) Ether
}

type Constants struct {
	ProjectID    string `mapstructure:"PROJECT_ID"`
	ProxyAddress string `mapstructure:"ALL_PROXY"`
}

func (c *Constants) Coalesce() Ether {
	// 1. Tell Viper where to look for the universal environment file
	viper.AddConfigPath(".")    // Search in the current working directory
	viper.SetConfigName(".env") // Look for a file named ".env"
	viper.SetConfigType("env")  // Treat the file as a .env format

	// 2. Read the file
	viper.ReadInConfig()

	// 3. Special Project Id handling to set PROJECT_ID back into the environment
	projectId := os.Getenv("PROJECT_ID")
	if projectId == "" {
		projectId = viper.GetString("PROJECT_ID")
	}

	if projectId != "" {
		LoadSecrets(&GCPSecretManager{ProjectID: projectId})
	}

	viper.BindEnv("ALL_PROXY")
	viper.BindEnv("PROJECT_ID")
	viper.AutomaticEnv()
	viper.Unmarshal(&c)

	return c
}

func (c *Constants) MergeFromFile(filename string) Ether {
	// 1. look for addition environment file
	viper.SetConfigFile(filename)
	_ = viper.MergeInConfig()
	viper.Unmarshal(c)
	return c
}

func (c *Constants) MergeFromCmd(cmd *cobra.Command) Ether {
	if flag := cmd.Flags().Lookup("project"); flag != nil {
		viper.BindPFlag("PROJECT_ID", flag)
	}

	viper.Unmarshal(c)
	return c
}
