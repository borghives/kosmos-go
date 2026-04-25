package ether

type Constants struct {
	ProjectID    string `mapstructure:"PROJECT_ID" cmdflag:"project,permeate"`
	ProxyAddress string `mapstructure:"ALL_PROXY" cmdflag:",permeate"`
	DebugLevel   int    `mapstructure:"DEBUG_LEVEL" cmdflag:"debug,permeate"`
}

var UniversalConstants LiminalStructure[Constants]

func CollapseUniversalConstants() Constants {
	return UniversalConstants.Collapse()
}
