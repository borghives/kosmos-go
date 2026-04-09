package kosmos

import (
	"log"

	"github.com/borghives/kosmos-go/ether"
	"github.com/borghives/kosmos-go/observation"
	"github.com/spf13/cobra"
)

// Ignite Base the kosmos where Secret Manager is optional and no Observer Client is created.
// Failure to do so will be fatal to the application.  An application cannot exist without the kosmos.
func IgniteBase(cmdSource *cobra.Command, source ...string) {
	ether.UniversalConstants.MergeFromFile(source...)
	if cmdSource != nil {
		ether.UniversalConstants.MergeFromCmd(cmdSource)
	}

	projectId := ether.UniversalConstants.Collapse().ProjectID
	if projectId != "" {
		if err := ether.LoadSecretsFile(&ether.GCPSecretManager{ProjectID: projectId}); err != nil {
			log.Fatalf("Fatal: Failed to load secrets file: %v", err)
		}
	}

	if err := ether.CollapseKnownEthers(cmdSource, source...); err != nil {
		log.Fatalf("Fatal: Failed to collapse known ethers: %v", err)
	}

}

// Ignite the kosmos.  Failure to do so will be fatal to the application.  An application cannot exist without the kosmos.
func Ignite(cmdSource *cobra.Command, source ...string) {
	ether.UniversalConstants.MergeFromFile(source...)
	if cmdSource != nil {
		ether.UniversalConstants.MergeFromCmd(cmdSource)
	}

	projectId := ether.UniversalConstants.Collapse().ProjectID
	if projectId == "" {
		log.Fatalf("Fatal: Failed to ignite universal constants: ProjectID")
	}

	if err := ether.LoadSecretsFile(&ether.GCPSecretManager{ProjectID: projectId}); err != nil {
		log.Fatalf("Fatal: Failed to load secrets file: %v", err)
	}

	if err := ether.CollapseKnownEthers(cmdSource, source...); err != nil {
		log.Fatalf("Fatal: Failed to collapse known ethers: %v", err)
	}

	MustHaveObserverClient()
}

func IsSecretSourceFormat(s string) bool {
	return ether.IsSecretSourceFormat(s)
}

func CollapseSecretString(s string) (string, error) {
	return ether.CollapseSecretString(s)
}

func SummonSecretManager() ether.SecretManager {
	return ether.SummonSecretManager()
}

func SummonObservationFor(purpose observation.PurposeAffinity) *observation.MongoDataverse {
	return observation.SummonMongo(purpose)
}
