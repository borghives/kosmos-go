# Antigravity Rules & Guidelines


Behavioral guidelines to reduce common LLM coding mistakes. Merge with project-specific instructions as needed.

**Tradeoff:** These guidelines bias toward caution over speed. For trivial tasks, use judgment.

## 1. Think Before Coding

**Don't assume. Don't hide confusion. Surface tradeoffs.**

Before implementing:
- State your assumptions explicitly. If uncertain, ask.
- If multiple interpretations exist, present them - don't pick silently.
- If a simpler approach exists, say so. Push back when warranted.
- If something is unclear, stop. Name what's confusing. Ask.

## 2. Simplicity First

**Minimum code that solves the problem. Nothing speculative.**

- No features beyond what was asked.
- No abstractions for single-use code.
- No "flexibility" or "configurability" that wasn't requested.
- No error handling for impossible scenarios.
- If you write 200 lines and it could be 50, rewrite it.

Ask yourself: "Would a senior engineer say this is overcomplicated?" If yes, simplify.

## 3. Surgical Changes

**Touch only what you must. Clean up only your own mess.**

When editing existing code:
- Don't "improve" adjacent code, comments, or formatting.
- Don't refactor things that aren't broken.
- Match existing style, even if you'd do it differently.
- If you notice unrelated dead code, mention it - don't delete it.

When your changes create orphans:
- Remove imports/variables/functions that YOUR changes made unused.
- Don't remove pre-existing dead code unless asked.

The test: Every changed line should trace directly to the user's request.

## 4. Goal-Driven Execution

**Define success criteria. Loop until verified.**

Transform tasks into verifiable goals:
- "Add validation" → "Write tests for invalid inputs, then make them pass"
- "Fix the bug" → "Write a test that reproduces it, then make it pass"
- "Refactor X" → "Ensure tests pass before and after"

For multi-step tasks, state a brief plan:
```
1. [Step] → verify: [check]
2. [Step] → verify: [check]
3. [Step] → verify: [check]
```

Strong success criteria let you loop independently. Weak criteria ("make it work") require constant clarification.

---

**These guidelines are working if:** fewer unnecessary changes in diffs, fewer rewrites due to overcomplication, and clarifying questions come before implementation rather than after mistakes.


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