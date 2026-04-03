package observation

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/borghives/kosmos-go/ether"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"golang.org/x/net/proxy"
)

type PurposeAffinity int

const (
	PurposeAffinityUnknown PurposeAffinity = iota
	PurposeAffinityAdmin
	PurposeAffinityCreator
	PurposeAffinityObserver
	PurposeAffinityCount //Max Count/Value for PurposeAffinity
)

func (p PurposeAffinity) String() string {
	switch p {
	case PurposeAffinityUnknown:
		return "Unknown"
	case PurposeAffinityAdmin:
		return "Admin"
	case PurposeAffinityCreator:
		return "Creator"
	case PurposeAffinityObserver:
		return "Observer"
	default:
		return "Unknown"
	}
}

func CollapseMongoURISecret(uri string) (string, error) {
	// 1. Isolate the scheme
	schemeSplit := strings.SplitN(uri, "://", 2)
	if len(schemeSplit) != 2 {
		return "", fmt.Errorf("invalid URI format")
	}
	scheme, remainder := schemeSplit[0], schemeSplit[1]

	// 2. Find the end of the credentials (the LAST '@' before any '/' or '?')
	// This is important because passwords themselves can contain '@' if encoded,
	// but the delimiter between creds and hosts is the final '@'.
	endOfCreds := strings.LastIndex(remainder, "@")
	if endOfCreds == -1 {
		return uri, nil // No credentials found (unauthenticated connection)
	}

	creds := remainder[:endOfCreds]
	hostAndPath := remainder[endOfCreds:] // Includes the '@'

	// 3. Split User and Password
	userAuth := strings.SplitN(creds, ":", 2)
	if len(userAuth) < 2 {
		return uri, nil // Only user, no password
	}

	user, pass := userAuth[0], userAuth[1]

	if user != "" {
		fmt.Printf("URI user: %s\n", user)
	}

	// 4. Translate and Stitch
	var err error
	if ether.IsSecretSource(pass) {
		pass, err = ether.CollapseSecretSource(pass)
		if err != nil {
			return "", err
		}
	}
	return fmt.Sprintf("%s://%s:%s%s", scheme, user, pass, hostAndPath), nil
}

func MaskMongoURI(uri string) (string, error) {
	// 1. Isolate the scheme
	schemeSplit := strings.SplitN(uri, "://", 2)
	if len(schemeSplit) != 2 {
		return "", fmt.Errorf("invalid URI format")
	}
	scheme, remainder := schemeSplit[0], schemeSplit[1]

	// 2. Find the end of the credentials (the LAST '@' before any '/' or '?')
	// This is important because passwords themselves can contain '@' if encoded,
	// but the delimiter between creds and hosts is the final '@'.
	endOfCreds := strings.LastIndex(remainder, "@")
	if endOfCreds == -1 {
		return uri, nil // No credentials found no need to mask
	}

	creds := remainder[:endOfCreds]
	hostAndPath := remainder[endOfCreds:] // Includes the '@'

	// 3. Split User and Password
	userAuth := strings.SplitN(creds, ":", 2)
	if len(userAuth) < 2 {
		return uri, nil // Only user, no password no need to mask
	}

	user, _ := userAuth[0], userAuth[1]

	return fmt.Sprintf("%s://%s:****%s", scheme, user, hostAndPath), nil
}

func CollapseMainDatabaseName() string {
	constants := ether.CollapseDataverseConstants()

	if constants.Database == "" {
		log.Fatalf("Main Database (MONGODB_DATABASE) is not set")
	}
	return constants.Database
}

func CollapseURIFor(purpose PurposeAffinity) (string, error) {
	constants := ether.CollapseDataverseConstants()
	if constants.CmdUri != "" {
		return CollapseMongoURISecret(constants.CmdUri)
	}

	switch purpose {
	case PurposeAffinityObserver:
		return CollapseMongoURISecret(constants.Uri)
	case PurposeAffinityCreator:
		return CollapseMongoURISecret(constants.CreatorUri)
	case PurposeAffinityAdmin:
		return CollapseMongoURISecret(constants.AdminUri)
	default:
		return constants.Uri, nil
	}
}

// key value element shorthand
func kv(key string, value any) bson.E {
	return bson.E{Key: key, Value: value}
}

type MongoDataverse struct {
	clientOption *options.ClientOptions
	client       *mongo.Client
	purpose      PurposeAffinity
}

type MongoRole struct {
	Role string `bson:"role"`
	DB   string `bson:"db"`
}

type MongoMember struct {
	User  string      `bson:"user"`
	Roles []MongoRole `bson:"roles"`
}

type MembersInfoResponse struct {
	Users []MongoMember `bson:"users"`
}

type MemberResponsibility struct {
	ReadDbs      []string
	ReadWriteDbs []string
	CreatorDbs   []string
	IsAdmin      bool
}

func (r *MemberResponsibility) ToMongoRoles() []MongoRole {
	var roles []MongoRole
	for _, db := range r.ReadDbs {
		roles = append(roles, MongoRole{Role: "read", DB: db})
	}
	for _, db := range r.ReadWriteDbs {
		roles = append(roles, MongoRole{Role: "readWrite", DB: db})
	}
	for _, db := range r.CreatorDbs {
		roles = append(roles, MongoRole{Role: "dbAdmin", DB: db})
		roles = append(roles, MongoRole{Role: "readWrite", DB: db})
	}
	if r.IsAdmin {
		roles = append(roles, MongoRole{Role: "userAdmin", DB: "admin"})
		roles = append(roles, MongoRole{Role: "clusterAdmin", DB: "admin"})
	}
	return roles
}

var (
	mongoObservers     [PurposeAffinityCount]*MongoDataverse
	mongoObserversOnce [PurposeAffinityCount]sync.Once
)

func SummonMongo(purpose PurposeAffinity) *MongoDataverse {
	mongoObserversOnce[purpose].Do(func() {
		clientOptions, err := coalesceMongoOptionsFor(purpose)
		if err != nil {
			log.Fatalf("Failed to coalesce mongo options for purpose %v: %v", purpose, err)
		}
		mongoObservers[purpose] = &MongoDataverse{clientOption: clientOptions, purpose: purpose}
	})
	return mongoObservers[purpose]
}

func (m *MongoDataverse) Client() *mongo.Client {
	if m.client == nil {
		var err error
		m.client, err = m.Connect()
		if err != nil {
			log.Fatalf("Failed to connect to MongoDB: %v", err)
		}
	}
	return m.client
}

func (m *MongoDataverse) Connect() (*mongo.Client, error) {
	client, err := tryConnectMongo(m.clientOption, 10)
	if err != nil {
		return nil, err
	}
	m.client = client
	return client, nil
}

func (m *MongoDataverse) Close() {
	if m.client != nil {
		err := m.client.Disconnect(context.Background())
		if err != nil {
			log.Printf("Failed to disconnect from MongoDB: %v", err)
		}
		m.client = nil
	}
}

func (m *MongoDataverse) ListMembers() (*MembersInfoResponse, error) {
	usersInfoCmd := bson.D{kv("usersInfo", 1)}

	var result MembersInfoResponse
	err := m.runAdministrativeCommand(usersInfoCmd).Decode(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (m *MongoDataverse) RemoveMember(name string) error {
	usersInfoCmd := bson.D{kv("dropUser", name)}

	var result bson.M
	err := m.runAdministrativeCommand(usersInfoCmd).Decode(&result)
	if err != nil {
		return err
	}

	return nil
}

func (m *MongoDataverse) UpdateMember(name string, newPassword string, responsibility MemberResponsibility, upsert bool) error {
	if newPassword == "" {
		return fmt.Errorf("Cannot set empty password\n")
	}

	roles := responsibility.ToMongoRoles()
	userCmd := bson.D{
		kv("updateUser", name),
		kv("pwd", newPassword),
		kv("roles", roles),
	}

	var result bson.M
	err := m.runAdministrativeCommand(userCmd).Decode(&result)
	if err != nil {
		if upsert {
			// try create user as part of upsert
			userCmd[0] = kv("createUser", name)
			err = m.runAdministrativeCommand(userCmd).Decode(&result)
		}
		return err
	}

	return nil
}

// RSStatus defines the top-level response from replSetGetStatus
type RSStatus struct {
	Set     string   `bson:"set"`
	Members []Member `bson:"members"`
}

// Member defines the individual node status in the replica set
type Member struct {
	Name           string    `bson:"name"`
	StateStr       string    `bson:"stateStr"`
	OptimeDate     time.Time `bson:"optimeDate"`
	SyncSourceHost string    `bson:"syncSourceHost"`
	Health         int       `bson:"health"` // 1 for up, 0 for down
}

type ServerStatus struct {
	Uptime      int64 `bson:"uptime"`
	Connections struct {
		Current   int `bson:"current"`
		Available int `bson:"available"`
	} `bson:"connections"`
	Opcounters struct {
		Insert  int64 `bson:"insert"`
		Query   int64 `bson:"query"`
		Update  int64 `bson:"update"`
		Delete  int64 `bson:"delete"`
		Command int64 `bson:"command"`
	} `bson:"opcounters"`
	Mem struct {
		Resident int64 `bson:"resident"` // in MB
		Virtual  int64 `bson:"virtual"`
	} `bson:"mem"`
}

type MongoStatus struct {
	RSStatus     RSStatus
	ServerStatus ServerStatus
}

func (m *MongoDataverse) Status() (*MongoStatus, error) {
	var status MongoStatus
	err := m.runAdministrativeCommand(bson.D{kv("replSetGetStatus", 1)}).Decode(&status.RSStatus)
	if err != nil {
		return nil, err
	}
	err = m.runAdministrativeCommand(bson.D{kv("serverStatus", 1)}).Decode(&status.ServerStatus)
	if err != nil {
		return nil, err
	}
	return &status, nil
}

func (m *MongoDataverse) ReelectPrimary() error {
	return m.runAdministrativeCommand(bson.D{kv("replSetStepDown", 60)}).Err()

}

func (m *MongoDataverse) AdminDatabase() *mongo.Database {
	return m.Client().Database("admin")
}

func (m *MongoDataverse) MainDatabase() *mongo.Database {
	return m.Client().Database(CollapseMainDatabaseName())
}

func (m *MongoDataverse) BranchDatabase(branchName string) *mongo.Database {
	if branchName == "" || branchName == "main" {
		return m.MainDatabase()
	}
	return m.Client().Database(branchName)
}

func (m *MongoDataverse) Collection(name string) *mongo.Collection {
	return m.MainDatabase().Collection(name)
}

func (m *MongoDataverse) runAdministrativeCommand(cmd bson.D) *mongo.SingleResult {
	return m.AdminDatabase().RunCommand(context.Background(), cmd)
}

type mongoDialerWrapper struct {
	dialer proxy.Dialer
}

func (m *mongoDialerWrapper) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	if cd, ok := m.dialer.(interface {
		DialContext(context.Context, string, string) (net.Conn, error)
	}); ok {
		return cd.DialContext(ctx, network, addr)
	}
	return m.dialer.Dial(network, addr)
}

func coalesceMongoOptionsFor(purpose PurposeAffinity) (*options.ClientOptions, error) {
	uri, err := CollapseURIFor(purpose)
	if err != nil {
		return nil, err
	}

	clientOptions := options.Client().ApplyURI(uri)

	proxyAddress := ether.CollapseConstants().ProxyAddress
	if proxyAddress != "" {
		log.Println("Using proxy for MongoDB: ", proxyAddress)
		proxyUrl, err := url.Parse(proxyAddress)
		if err != nil {
			return nil, fmt.Errorf("Failed to parse proxy address: %v", err)
		}

		dialer, err := proxy.FromURL(proxyUrl, proxy.Direct)
		if err != nil {
			return nil, fmt.Errorf("Failed to create dialer: %v", err)
		}

		clientOptions = clientOptions.SetDialer(&mongoDialerWrapper{dialer: dialer})
	}
	return clientOptions, nil
}

func tryConnectMongo(clientOptions *options.ClientOptions, n int) (*mongo.Client, error) {
	var client *mongo.Client
	var err error
	for i := range n {
		client, err = mongo.Connect(clientOptions)
		if err == nil {
			err = client.Ping(context.Background(), nil)
		}
		if err == nil {
			log.Printf("MongoDb Ping Success")
			return client, nil
		}
		log.Printf("MongoDb Ping Failed.  Waiting for MongoDB... (attempt %d): %v", i+1, err)
		time.Sleep(5 * time.Second)
	}

	return nil, err
}
