package kosmos

import (
	"github.com/borghives/kosmos-go/ether"
	"github.com/borghives/kosmos-go/observation"
)

func IsSecretSource(s string) bool {
	return ether.IsSecretSource(s)
}

func CollapseSecret(s string) (string, error) {
	return ether.CollapseSecret(s)
}

func SummonSecretManager() ether.SecretManager {
	return ether.SummonSecretManager()
}

func SummonDataverseFor(purpose observation.PurposeAffinity) *observation.MongoDataverse {
	return observation.SummonMongo(purpose)
}
