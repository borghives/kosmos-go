# Kosmos-Go

**Kosmos-Go** is a Go-based framework and data persistence layer built around MongoDB.

## The Microcosm Philosophy

The naming strategy in `kosmos-go` treats the database as the foundational fabric of reality. Operations involve resolving probability into concrete data, observing state, and managing causal side-effects.

### Foundational Terminology

- **Collapse / Collapsable**: The transition from potential to actual state. Used for resolving database IDs, establishing timestamps, and pulling secrets into tangible values.
- **Witness / Observer**: The act of fixing an entity's state into empirical reality. An `Observer` interacts with MongoDB, and to `Witness` means to persist a specific state.
- **Ripple**: Causal side-effects extending outwards from an event (used for defining hooks or secondary reactive operations alongside a `Witness` or `Collapse`).
- **Entangled**: A state of being linked to the fabric of reality (i.e., whether an entity currently exists with an ID in the database).
- **Summon**: Calling forth an authoritative entity, factory, or singleton (e.g., `SummonSecretManager`, `SummonMongo`).
- **Coalesce**: Merging unformed configuration data, environment variables, and connections into a unified single source of truth.
- **Ether**: The ambient layer handling environment variables, configuration streams, and secrets.
- **Observable**: Stateful structural interfaces representing objects that can be witnessed, filtered, or tracked.

## Architecture & Packages

### `kosmos`
The top-level package that exposes the core capabilities: summoning observers and secretly collapsing strings using the ether layer.

### `model`
The data representation layer. Models embedded with `kosmos.BaseModel` gain database capabilities via custom struct tags (`kdb` for database name, `kcol` for collection name). 
- **Operations:** Use commands like `kosmos.Filter`, `kosmos.Witness`, and `kosmos.All` to manipulate your data.
- **Lifecycle:** As a model's state converges, the `BaseModel` automatically defines fields like `_id`, `updated_time`, and `created_time` under `Collapse()` events.

### `observation`
The data connectivity layer bridging the application to MongoDB.
- **MongoObserver**: Maintains robust pooling and connections with different `PurposeAffinity` roles (e.g., `Admin`, `Creator`, `Observer`), allowing specialized access rights.
- Supports administrative commands, proxy connections, and replica set status management.

### `ether`
The foundational configuration and secrets manager.
- Integrates with Google Secret Manager (`cloud.google.com/go/secretmanager`) and keychain adapters.
- Evaluates raw URI strings and secrets dynamically, allowing credentials to be injected transparently into connections without hardcoding.

## Getting Started

1. Set up your MongoDB credentials using the `ether` package.
2. Initialize models by embedding `kosmos.BaseModel` and assigning `kdb` and `kcol` tags.
3. `Summon` your observers (`observation.MongoObserver`) with an intended affinity.
4. Begin `Witness`ing and `Filter`ing reality!