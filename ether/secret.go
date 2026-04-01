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
	"time"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"google.golang.org/api/iterator"

	"github.com/zalando/go-keyring"
)

func IsSecretSource(s string) bool {
	return strings.HasPrefix(string(s), "__secret:") && strings.HasSuffix(string(s), "__")
}

func CollapseSecret(s string) (string, error) {
	if IsSecretSource(s) {
		return CollapseSecretSource(s)
	}
	parts := strings.Split(s, ":")

	if len(parts) == 2 {
		return SummonSecretManager().AccessSecret(parts[0], parts[1])
	}
	return SummonSecretManager().AccessSecret(s, "latest")
}

func CollapseSecretSource(s string) (string, error) {
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
	secret, err := SummonSecretManager().AccessSecret(parts[1], parts[2])
	if err != nil {
		return "", err
	}
	if secret == "" {
		return "", fmt.Errorf("Secret from Secret Manager is empty")
	}
	return secret, nil
}

// SecretManager is an interface that allows fetching secrets from different backends.
type SecretManager interface {
	AccessSecret(name string, version string) (string, error)
	ListSecrets() ([]SecretInfo, error)
	CreateSecret(name string) error
	AddSecretVersion(name, payload string) error
	IsSecretStale(name string, ttlHour int) bool
}

// Load parses a .envsecret file and then loads all the variables found as environment variables.
// It uses the provided SecretManager to fetch the actual secret value.
func LoadSecrets(manager SecretManager, filenames ...string) error {
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
		err := loadFile(manager, filename)
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

func (m *GCPSecretManager) AccessSecret(name string, version string) (string, error) {
	if m.ProjectID == "" {
		return "", errors.New("Project ID is missing. Set GOOGLE_CLOUD_PROJECT or PROJECT_ID environment variable.")
	}

	ctx := context.Background()

	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create secretmanager client: %w", err)
	}
	defer client.Close()

	location := fmt.Sprintf("projects/%s/secrets/%s/versions/%s", m.ProjectID, name, version)
	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: location,
	}

	result, err := client.AccessSecretVersion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to access secret version: %w", err)
	}

	return string(result.Payload.Data), nil
}

type SecretInfo struct {
	Name    string
	Version string
}

func (m *GCPSecretManager) ListSecrets() ([]SecretInfo, error) {
	projectParent := fmt.Sprintf("projects/%s", m.ProjectID)
	// 1. Build the request to list secrets
	req := &secretmanagerpb.ListSecretsRequest{
		Parent: projectParent,
	}

	ctx := context.Background()

	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create secretmanager client: %w", err)
	}
	defer client.Close()

	var secrets []SecretInfo
	iter := client.ListSecrets(ctx, req)
	for {
		secret, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list secrets: %w", err)
		}
		secretParts := strings.Split(secret.Name, "/")

		if len(secretParts) < 4 {
			return nil, fmt.Errorf("failed to list secrets: %s", secret.Name)
		}
		name := secretParts[3]
		var version string
		if len(secretParts) > 4 {
			version = secretParts[4]
		}

		secrets = append(secrets, SecretInfo{Name: name, Version: version})
	}

	return secrets, nil
}

func (m *GCPSecretManager) CreateSecret(name string) error {
	projectParent := fmt.Sprintf("projects/%s", m.ProjectID)
	// 1. Create the Secret (Metadata container)
	createSecretReq := &secretmanagerpb.CreateSecretRequest{
		Parent:   projectParent,
		SecretId: name,
		Secret: &secretmanagerpb.Secret{
			Replication: &secretmanagerpb.Replication{
				Replication: &secretmanagerpb.Replication_Automatic_{
					Automatic: &secretmanagerpb.Replication_Automatic{},
				},
			},
		},
	}

	ctx := context.Background()

	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create secretmanager client: %w", err)
	}
	defer client.Close()

	_, err = client.CreateSecret(ctx, createSecretReq)
	return err
}

func (m *GCPSecretManager) AddSecretVersion(name, payload string) error {
	addSecretVersionReq := &secretmanagerpb.AddSecretVersionRequest{
		Parent: fmt.Sprintf("projects/%s/secrets/%s", m.ProjectID, name),
		Payload: &secretmanagerpb.SecretPayload{
			Data: []byte(payload),
		},
	}

	ctx := context.Background()

	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create secretmanager client: %w", err)
	}
	defer client.Close()

	_, err = client.AddSecretVersion(ctx, addSecretVersionReq)
	return err
}

func (m *GCPSecretManager) IsSecretStale(name string, ttlHour int) bool {
	req := &secretmanagerpb.GetSecretVersionRequest{
		Name: fmt.Sprintf("projects/%s/secrets/%s/versions/latest", m.ProjectID, name),
	}

	ctx := context.Background()

	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		log.Printf("failed to create secretmanager client: %v", err)
		return false
	}
	defer client.Close()

	version, err := client.GetSecretVersion(ctx, req)
	if err != nil {
		log.Printf("failed to get secret version: %v", err)
		return false
	}

	return version.CreateTime.AsTime().Before(time.Now().Add(-time.Duration(ttlHour) * time.Hour))
}

type LocalKeyring struct{}

func NewLocalKeyring() *LocalKeyring {
	return &LocalKeyring{}
}

func (m *LocalKeyring) AccessSecret(secretID, versionID string) (string, error) {
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

func (m *LocalKeyring) ListSecrets() ([]SecretInfo, error) {
	return nil, errors.New("ListSecrets not supported for LocalKeyring")
}

func (m *LocalKeyring) CreateSecret(name string) error {
	return errors.New("CreateSecret not supported for LocalKeyring")
}

func (m *LocalKeyring) AddSecretVersion(name, payload string) error {
	return errors.New("AddSecretVersion not supported for LocalKeyring")
}

func (m *LocalKeyring) IsSecretStale(name string, ttlHour int) bool {
	return false
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

func loadFile(manager SecretManager, filename string) error {
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
			secretValue, err := manager.AccessSecret(secretName, versionID)
			if err != nil {
				fmt.Printf("Warning: Failed to load secret '%s' for environment variable '%s': %v\n", secretID, envVar, err)
				continue
			}
			os.Setenv(envVar, secretValue)
		}
	}

	return scanner.Err()
}
