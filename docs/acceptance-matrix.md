# Acceptance matrix

| Case | Why it is canonical | Expected | Evidence preserved |
|---|---|---|---|
| normal-independent | A and B write disjoint receipt fields | CLOSED | two equal IR/artifact/provenance triples |
| normal-idempotent | A and B write the same value | CLOSED | two equal IR/artifact/provenance triples |
| normal-nested | A and B write disjoint nested fields | CLOSED | two equal IR/artifact/provenance triples |
| unknown-missing-dependency | B names an absent dependency | UNKNOWN | dependency frontier and six UNKNOWN fields |
| unknown-missing-scope | A has no declared scope | UNKNOWN | operation scope frontier and six UNKNOWN fields |
| unknown-missing-input | baseline omits receipt.total | UNKNOWN | input frontier and six UNKNOWN fields |
| refuted-order-dependent | last writer changes between A→B and B→A | REFUTED | minimal path counterexample and both order digests |
| refuted-authority-conflict | incompatible authorities contend on one path | REFUTED | authority frontier and counterexample |
| refuted-guardrail-conflict | an operation writes a forbidden path | REFUTED | guardrail frontier and counterexample |

The CI job fixes the denominator before evaluating fixtures and records integer
counts only. The repository merge authority and runtime source-write authority
are zero.
