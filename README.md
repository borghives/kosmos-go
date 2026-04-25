# Kosmos-Go

**Kosmos-Go** is a Go-based framework and data persistence layer built around MongoDB.

## The Microcosm Philosophy

The naming strategy in `kosmos-go` treats the database as foundational. 

### Foundational Terminology

- **Collapse / Collapsable**: The transition from potential to actual state. Used for resolving database IDs, establishing timestamps, and pulling secrets into tangible values.
- **Witness / Observer**: The act of fixing an entity's state into empirical reality. An `Observer` interacts with MongoDB, and to `Witness` means to persist a specific state.
- **Ripple**: Causal side-effects extending outwards from an event (used for defining hooks or secondary reactive operations alongside a `Witness` or `Collapse`).
- **Entangled**: A state of being linked to the fabric of reality (i.e., whether an entity currently exists with an ID in the database).
- **Summon**: Calling forth an authoritative entity, factory, or singleton (e.g., `SummonSecretManager`, `SummonMongo`).
- **Coalesce**: Merging unformed configuration data, environment variables, and connections into a unified single source of truth.
- **Ether**: The ambient layer handling environment variables, configuration streams, and secrets.
- **Detectable**: Interfaces representing objects that can be witnessed, filtered, or tracked.

## Architecture & Packages

### `kosmos`
The top-level package that exposes the core capabilities:
- Setting up the ambient environment via `Ignite` and `IgniteBase`.
- The data representation layer with `kosmos.BaseModel`. Models embedded with `kosmos.BaseModel` gain database capabilities via a custom struct tag (`kosmos:"branch>collection"` or `kosmos:"collection"`).
- **Operations:** Use commands like `kosmos.Detect`, `kosmos.Witness`, and `kosmos.All` to manipulate your data.
- **Lifecycle:** As a model's state converges, the `BaseModel` automatically defines fields like `_id`, `updated_time`, and `created_time` under `Collapse()` events.

### `observation`
The data connectivity layer bridging the application to MongoDB.
- **EntityObserver & EntityDetector**: Used internally by `kosmos.Witness` and `kosmos.Detect` to interact with collections.
- **MongoDataverse**: Maintains robust pooling and connections with different `PurposeAffinity` roles (e.g., `Admin`, `Creator`, `Observer`), allowing specialized access rights.
- Supports administrative commands, proxy connections, and replica set status management.

### `ether`
The foundational configuration and secrets manager.
- Integrates with Google Secret Manager (`cloud.google.com/go/secretmanager`) and `viper`/`cobra` for command-line flags and environment variables.
- Uses `Liminal` interfaces to manage configuration structures securely.
- Evaluates raw URI strings and secrets dynamically, allowing credentials to be injected transparently into connections without hardcoding.

## Getting Started

1. Initialize your environment by calling `kosmos.Ignite` or `kosmos.IgniteBase`, which will load environment variables and secrets via the `ether` package.
2. Initialize models by embedding `kosmos.BaseModel` as the first field and assigning the `kosmos` struct tag to map to a branch and collection.
```go
type MyModel struct {
    kosmos.BaseModel `bson:",inline" kosmos:"my_branch>my_collection"`
    Name string `bson:"name"`
}
```
3. Begin `Witness`ing and `Detect`ing reality!
```go
obj := &MyModel{Name: "Hello World"}
err := kosmos.Witness(ctx, obj)
```