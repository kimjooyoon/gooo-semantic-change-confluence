# gooo-semantic-change-confluence

This repository closes one narrow Gooo capability: deciding whether two
independent semantic changes can be applied in either order without changing
meaning. The semantic contract lives in
[`.gooo/confluence.gooo`](.gooo/confluence.gooo); Go is only the interpreter, generator, verifier, and CI
measurement runtime.

The source language is `.gooo`. A source input declares a baseline semantic
state and two change operations. The evaluator constructs a semantic graph,
applies `change-a → change-b` and `change-b → change-a`, emits canonical
semantic IR, a derived Go artifact, and canonical provenance for each order,
then compares all three digests. Only equality of all three produces
`CONFLUENT/CLOSED`. A missing frontier is `UNKNOWN`; order dependence or an
authority/guardrail conflict is `REFUTED`. UNKNOWN always carries stage, step,
reason, unknown_class, next_operation, and blocked_by.

The executable example is
[`examples/parallel-receipt-changes.gooo`](examples/parallel-receipt-changes.gooo):
two receipt changes are generated in both orders, the generated Go is executed,
and an independent verifier checks the result. CI also runs nine fixed cases:
three normal, three UNKNOWN, and three REFUTED, reduced in the declared order
`REFUTED > UNKNOWN > CLOSED`.

The product never merges repositories or writes source inputs:
`merge_authority=0`, `runtime_write=0`, `cross_project_required_gates=0`, and
generated files are written only below caller-owned output directories. GitHub
Actions is the verification authority. The root README is excluded from
inventory counts. External user utility is `UNKNOWN` because no external-user
evidence is bundled.

## Evidence

Every CI run uploads a `semantic-confluence-<sha>` artifact containing the
human-readable conformance report, per-case results, both order bundles for
closed cases, generated-artifact execution evidence, and exact integer metrics.
The metrics include compile/build/test/conformance/integration wall time and
peak RSS, test total/selected/executed/reused/failed/unknown, Go and Gooo file
counts and physical lines, regular files, descendant directories, outputs and
generated artifact counts/bytes, local/remote authority counts, and the
preserved `local_validation_commands=0` plus `OPERATIONAL_REFUTED` marker.

The release workflow is main-bound and refuses existing tags/releases. It
requires GitHub's immutable-releases repository setting, creates one annotated
tag, publishes the exact source/binary/report/manifest/SHA256SUMS asset set,
and verifies release immutability, release ID, tag object and peeled target,
asset IDs, sizes, and digests through the GitHub API.
