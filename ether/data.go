package ether

type MongoConstants struct {
	Uri        string `mapstructure:"MONGODB_URI" cmdflag:"uri,permeate" `
	AdminUri   string `mapstructure:"MONGODB_ADMIN_URI" `
	CreatorUri string `mapstructure:"MONGODB_CREATOR_URI" `
	Database   string `mapstructure:"MONGODB_DATABASE" cmdflag:"database,permeate" `
}

var MongoDataverseConstants EtherStructure[MongoConstants]

func CollapseDataverseConstants() MongoConstants {
	return MongoDataverseConstants.Collapse()
}

func init() {
	RegisterEther(&MongoDataverseConstants)
}
