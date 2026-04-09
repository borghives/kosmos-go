# Antigravity Rules & Guidelines

## 1. Naming Strategy: The Quantum Engine
When generating code, defining data structures, or architecting systems for the "Kosmos" project, you MUST adhere to the following philosophical and structural naming conventions. The paradigm focuses on quantum mechanics, observation, and reality manifestation.

### Foundational Definitions (The Blueprint)
Use the following concepts and terminology when interacting with the state, lifecycle, and structure of objects:

* **`Collapse` / `Collapsable`**
  * **Concept:** Collapsing a quantum probability into a defined reality.
  * **Usage:** Use for resolving state, generating IDs, establishing creation times, or resolving secrets into actual tangible values. It represents the transition from potential to actual state.

* **`Witness` / `Observer`**
  * **Concept:** The act of observation that fixes an entity's state into empirical reality.
  * **Usage:** Use for database operations, persistence layers, or actors that track an entity's state. An `Observer` interacts with the datastore, and to `Witness` is to commit or persist a specific state.

* **`Ripple`**
  * **Concept:** Causal side effects extending outwards through reality from an event.
  * **Usage:** Use for defining reactive changes or secondary updates that must occur alongside a `Collapse` or `Witness` event (e.g., defining side-effects like MongoDB `$setOnInsert` operations).

* **`Decohered`**
  * **Concept:** The state of being inextricably linked to the underlying fabric of reality (the database).
  * **Usage:** Use for checking if an object has identity or exists in the system (e.g., `HasID()` to check if an entity already has a primary key/ID).

* **`Summon`**
  * **Concept:** Calling forth an authoritative entity or service into the active context.
  * **Usage:** Use for factory functions, singletons, or initialization methods that bring managers, configurations, or observers into the operational context (e.g., `SummonSecretManager()`).

* **`Coalesce`**
  * **Concept:** Bringing disparate pieces of unformed data into a unified, coherent whole.
  * **Usage:** Use for configuration builders, merging environment variables, files, and command-line arguments into a single structural source of truth.

* **`Ether`**
  * **Concept:** The ambient, all-permeating medium holding the configurations and secrets of the universe.
  * **Usage:** Use for the foundational packages or interfaces dealing with environment variables, secrets, configuration streams, and system boundaries.

* **`Observable`**
  * **Concept:** A stateful entity that exists and can be tracked, queried, or pulled upon.
  * **Usage:** Use for foundational model interfaces representing structs that can be resolved, filtered, or fetched from the reality storage.