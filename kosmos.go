package kosmos

import (
	"log"

	"github.com/borghives/kosmos-go/ether"
	"github.com/borghives/kosmos-go/observation"
)

// Ignite the kosmos.  Failure to do so will be fatal to the application.  An application cannot exist without the kosmos.
func Ignite(source ...string) {
	ether.UniversalConstants.MergeFromFile(source...)

	projectId := ether.UniversalConstants.Collapse().ProjectID
	if projectId == "" {
		log.Fatalf("Fatal: Failed to ignite universal constants: ProjectID")
	}

	if err := ether.LoadSecretsFile(&ether.GCPSecretManager{ProjectID: projectId}); err != nil {
		log.Fatalf("Fatal: Failed to load secrets file: %v", err)
	}

	if err := ether.CollapseKnownEthers(source...); err != nil {
		log.Fatalf("Fatal: Failed to collapse known ethers: %v", err)
	}

	MustHaveObserverClient()
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
