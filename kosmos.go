package kosmos

import (
	"context"
	"fmt"
	"strings"

	"github.com/borghives/kosmos-go/ether"
	"github.com/borghives/kosmos-go/observer"
)

func IsSecretSource(s string) bool {
	return strings.HasPrefix(string(s), "__secret:") && strings.HasSuffix(string(s), "__")
}

func CollapseSecret(s string) (string, error) {
	//parse string "__secret:name:version__"

	//check if the string is a holder string
	if !IsSecretSource(s) {
		return "", fmt.Errorf("Not a secret source string")
	}

	//remove the underscores
	s = s[2 : len(s)-2]

	//split the string by ":"
	parts := strings.Split(s, ":")

	//check if the string is a source string
	if len(parts) != 3 {
		return "", fmt.Errorf("Invalid secret source string format expect __secret:<name>:<version>__")
	}

	if parts[0] != "secret" {
		return "", fmt.Errorf("Invalid secret source string format expect __secret:<name>:<version>__")
	}

	//return the secret
	return ether.SummonSecretManager().AccessSecret(context.Background(), parts[1], parts[2])
}

func SummonObserverFor(purpose observer.PurposeAffinity) *observer.MongoObserver {
	return observer.SummonMongo(purpose)
}
