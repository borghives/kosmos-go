package ether

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/user"
	"strings"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"

	"github.com/zalando/go-keyring"
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
	return SummonSecretManager().AccessSecret(context.Background(), parts[1], parts[2])
}

// SecretManager is an interface that allows fetching secrets from different backends.
type SecretManager interface {
	AccessSecret(ctx context.Context, secretID, versionID string) (string, error)
}

// Load parses a .envsecret file and then loads all the variables found as environment variables.
// It uses the provided SecretManager to fetch the actual secret value.
func LoadSecrets(ctx context.Context, manager SecretManager, filenames ...string) error {
	if manager == nil {
		return errors.New("SecretManager is nil")
	}

	if loadDotenvsecretDisabled() {
		log.Println("dotenvsecret: .envsecret loading disabled by DOTENVSECRET_DISABLED environment variable")
		return nil
	}

	if len(filenames) == 0 {
		filenames = []string{".envsecret"}
	}

	for _, filename := range filenames {
		err := loadFile(ctx, manager, filename)
		if err != nil {
			return err
		}
	}
	return nil
}

func SummonSecretManager() SecretManager {
	constants := CollapseConstants()
	if constants.ProjectID == "" {
		log.Fatalf("Project ID is missing. Set GOOGLE_CLOUD_PROJECT or PROJECT_ID environment variable.")
	}
	return &GCPSecretManager{ProjectID: constants.ProjectID}
}

type GCPSecretManager struct {
	ProjectID string
}

func (m *GCPSecretManager) AccessSecret(ctx context.Context, secretID, versionID string) (string, error) {

	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		projectID = os.Getenv("PROJECT_ID")
	}
	if projectID == "" {
		return "", errors.New("Project ID is missing. Set GOOGLE_CLOUD_PROJECT or PROJECT_ID environment variable.")
	}

	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create secretmanager client: %w", err)
	}
	defer client.Close()

	name := fmt.Sprintf("projects/%s/secrets/%s/versions/%s", projectID, secretID, versionID)
	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: name,
	}

	result, err := client.AccessSecretVersion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to access secret version: %w", err)
	}

	return string(result.Payload.Data), nil
}

type LocalKeyring struct{}

func NewLocalKeyring() *LocalKeyring {
	return &LocalKeyring{}
}

func (m *LocalKeyring) AccessSecret(ctx context.Context, secretID, versionID string) (string, error) {
	username := os.Getenv("LOCAL_KEYRING_USERNAME")
	if username == "" {
		u, err := user.Current()
		if err == nil {
			username = u.Username
		} else {
			username = "default" // fallback if user.Current() fails
		}
	}

	secret, err := keyring.Get(secretID, username)
	if err != nil {
		return "", err
	}
	return secret, nil
}

// -- helper function --

func loadDotenvsecretDisabled() bool {
	val, ok := os.LookupEnv("DOTENVSECRET_DISABLED")
	if !ok {
		return false
	}
	val = strings.ToLower(val)
	return val == "1" || val == "true" || val == "t" || val == "yes" || val == "y"
}

func loadFile(ctx context.Context, manager SecretManager, filename string) error {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return nil
	}

	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			envVar := strings.TrimSpace(parts[0])
			secretID := strings.TrimSpace(parts[1])

			// Optional: strip quotes if present
			if (strings.HasPrefix(secretID, "\"") && strings.HasSuffix(secretID, "\"")) ||
				(strings.HasPrefix(secretID, "'") && strings.HasSuffix(secretID, "'")) {
				secretID = secretID[1 : len(secretID)-1]
			}

			secretParts := strings.Split(secretID, ":")
			secretName := secretParts[0]
			versionID := "latest"
			if len(secretParts) == 2 {
				versionID = secretParts[1]
			}

			// In python version: version_id is "latest", source_id is None by default
			secretValue, err := manager.AccessSecret(ctx, secretName, versionID)
			if err != nil {
				fmt.Printf("Warning: Failed to load secret '%s' for environment variable '%s': %v\n", secretID, envVar, err)
				continue
			}
			os.Setenv(envVar, secretValue)
		}
	}

	return scanner.Err()
}
