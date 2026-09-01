# Semantic change confluence protocol v1

The denominator is fixed before evaluation at nine canonical cases:

| Class | Cases | Required terminal |
|---|---:|---|
| normal | 3 | CLOSED |
| UNKNOWN | 3 | UNKNOWN |
| REFUTED | 3 | REFUTED |

The evaluator uses the semantic graph and two declared application orders in
`.gooo/confluence.gooo`. It binds each run to the same source, contract,
toolchain, and runner. A `CLOSED` result requires identical canonical semantic
IR bytes, generated artifact bytes, and provenance bytes for both orders.

`UNKNOWN` is fail-closed and must expose all six frontier fields:
`stage`, `step`, `reason`, `unknown_class`, `next_operation`, and `blocked_by`.
An unrecognized top-level decision is never inferred as `CLOSED` or
`FIXED_POINT`. `REFUTED` preserves a minimal counterexample frontier for order
dependence, authority conflict, or guardrail conflict.

This repository reports decisions and integer evidence. It does not emit a
score, percentage, weighted sum, or utility claim. The evaluator's merge and
runtime write authorities are both zero.
