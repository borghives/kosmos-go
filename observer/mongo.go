package observer

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/url"
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
	PurposeAffinityObserver
	PurposeAffinityCreator
	PurposeAffinityAdmin
	PurposeAffinityCount //Max Count/Value for PurposeAffinity
)

func CollapseURIFor(purpose PurposeAffinity) string {
	constants := ether.ColapseObserverConstants()
	switch purpose {
	case PurposeAffinityObserver:
		return constants.Uri
	case PurposeAffinityCreator:
		return constants.CreatorUri
	case PurposeAffinityAdmin:
		return constants.AdminUri
	default:
		return constants.Uri
	}
}

// key value element shorthand
func kv(key string, value any) bson.E {
	return bson.E{Key: key, Value: value}
}

type MongoObserver struct {
	clientOption *options.ClientOptions
	client       *mongo.Client
	purpose      PurposeAffinity
}

type MongoRole struct {
	Role string `bson:"role"`
	DB   string `bson:"db"`
}

type MongoUser struct {
	User  string      `bson:"user"`
	Roles []MongoRole `bson:"roles"`
}

type UsersInfoResponse struct {
	Users []MongoUser `bson:"users"`
}

type UserResponsibility struct {
	ReadDbs      []string
	ReadWriteDbs []string
	CreatorDbs   []string
	IsAdmin      bool
}

func (r *UserResponsibility) ToMongoRoles() []MongoRole {
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
	mongoObservers     [PurposeAffinityCount]*MongoObserver
	mongoObserversOnce [PurposeAffinityCount]sync.Once
)

func CoalesceMongoObserver(purpose PurposeAffinity) *MongoObserver {
	mongoObserversOnce[purpose].Do(func() {
		clientOptions, err := coalesceMongoOptionsFor(purpose)
		if err != nil {
			log.Fatalf("Failed to coalesce mongo options: %v", err)
		}
		mongoObservers[purpose] = &MongoObserver{clientOption: clientOptions, purpose: purpose}
	})
	return mongoObservers[purpose]
}

func (m *MongoObserver) Client() *mongo.Client {
	if m.client == nil {
		var err error
		m.client, err = m.Connect()
		if err != nil {
			log.Fatalf("Failed to connect to MongoDB: %v", err)
		}
	}
	return m.client
}

func (m *MongoObserver) Connect() (*mongo.Client, error) {
	client, err := tryConnectMongo(m.clientOption, 10)
	if err != nil {
		return nil, err
	}
	m.client = client
	return client, nil
}

func (m *MongoObserver) Close() {
	if m.client != nil {
		err := m.client.Disconnect(context.Background())
		if err != nil {
			log.Printf("Failed to disconnect from MongoDB: %v", err)
		}
		m.client = nil
	}
}

func (m *MongoObserver) QueryMembers() (*UsersInfoResponse, error) {
	usersInfoCmd := bson.D{kv("usersInfo", 1)}

	var result UsersInfoResponse
	err := m.runAdministrativeCommand(usersInfoCmd).Decode(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (m *MongoObserver) RemoveMember(name string) error {
	usersInfoCmd := bson.D{kv("dropUser", name)}

	var result bson.M
	err := m.runAdministrativeCommand(usersInfoCmd).Decode(&result)
	if err != nil {
		return err
	}

	return nil
}

func (m *MongoObserver) UpdateMember(name string, newPassword string, responsibility UserResponsibility, upsert bool) error {
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

func (m *MongoObserver) runAdministrativeCommand(cmd bson.D) *mongo.SingleResult {
	return m.Client().Database("admin").RunCommand(context.Background(), cmd)
}

func coalesceMongoOptionsFor(purpose PurposeAffinity) (*options.ClientOptions, error) {
	clientOptions := options.Client().ApplyURI(
		CollapseURIFor(purpose),
	)

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
