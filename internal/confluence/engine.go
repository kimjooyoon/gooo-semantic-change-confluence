package confluence

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
)

func LoadSpec(path string) (Spec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Spec{}, err
	}
	var spec Spec
	if err := decode(data, &spec); err != nil {
		return Spec{}, err
	}
	return spec, nil
}

func LoadInput(path string) (SourceInput, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return SourceInput{}, err
	}
	var input SourceInput
	if err := decode(data, &input); err != nil {
		return SourceInput{}, err
	}
	return input, nil
}

func Evaluate(spec Spec, input SourceInput) (Result, map[string]OrderEvidence, map[string]struct{ /* semantic ir, generated, provenance */ }) {
	result := Result{
		Schema:         "gooo/semantic-confluence-result/v1",
		CaseID:         input.CaseID,
		MergeAuthority: spec.Authorities.MergeAuthority,
	}
	artifacts := make(map[string]struct{})

	if reason, blocked, preflightDecision := preflight(spec, input); reason != "" {
		if preflightDecision == Refuted {
			result.Decision = Refuted
			result.SemanticVerdict = "NON_CONFLUENT"
			result.Counterexample = &Counterexample{
				Kind:     "guardrail-conflict",
				Reason:   reason,
				Frontier: sortedStrings(blocked),
			}
			return result, nil, artifacts
		}
		if blocked == nil {
			blocked = []string{"semantic:preflight"}
		}
		result.Decision = Unknown
		result.SemanticVerdict = "CONFLUENCE_UNKNOWN"
		result.Unknown = &UnknownEvidence{
			Stage:         "preflight",
			Step:          "validate-input-and-dependencies",
			Reason:        reason,
			UnknownClass:  "INSUFFICIENT_INPUT_OR_SCOPE",
			NextOperation: "supply the blocked semantic input and rerun both orders",
			BlockedBy:     sortedStrings(blocked),
		}
		return result, nil, artifacts
	}

	inputOperations := operationMap(input.Operations)
	specOperations := operationMap(spec.ChangeOperations)
	for _, pair := range [][]string{{"change-a", "change-b"}} {
		if reason, conflict := operationConflict(spec, inputOperations[pair[0]], inputOperations[pair[1]]); conflict {
			result.Decision = Refuted
			result.SemanticVerdict = "NON_CONFLUENT"
			result.Counterexample = &Counterexample{
				Kind:        "authority-conflict",
				Reason:      reason,
				Frontier:    []string{"operation:" + pair[0], "operation:" + pair[1]},
				Authorities: []string{inputOperations[pair[0]].Authority, inputOperations[pair[1]].Authority},
			}
			return result, nil, artifacts
		}
	}

	orders := make(map[string]OrderEvidence)
	orderNames := []string{"A_then_B", "B_then_A"}
	for _, orderName := range orderNames {
		order := spec.ApplicationOrders[orderName]
		state := copyState(input.Baseline)
		for _, operationID := range order {
			operation := inputOperations[operationID]
			for _, patch := range operation.Patches {
				state[patch.Path] = patch.Value
			}
		}
		canonicalOperations := make([]Operation, 0, len(specOperations))
		for _, operationID := range sortedOperationIDs(specOperations) {
			canonicalOperations = append(canonicalOperations, inputOperations[operationID])
		}
		ir := canonicalIR(input, spec, state, canonicalOperations)
		irDigest, irBytes, err := formatDigest(ir)
		if err != nil {
			return Result{}, nil, artifacts
		}
		generated := []byte(generatedSource(state))
		generatedDigest := digestBytes(generated)
		canonicalProvenance := provenance(canonicalOperations)
		provenanceDigest, provenanceBytes, err := formatDigest(canonicalProvenance)
		if err != nil {
			return Result{}, nil, artifacts
		}
		orders[orderName] = OrderEvidence{
			ExecutionOrder:         append([]string(nil), order...),
			SemanticIRFile:          filepath.Join(orderName, "semantic-ir.json"),
			SemanticIRDigest:        irDigest,
			GeneratedArtifactFile:   filepath.Join(orderName, "generated.go"),
			GeneratedArtifactDigest: generatedDigest,
			ProvenanceFile:          filepath.Join(orderName, "provenance.json"),
			ProvenanceDigest:        provenanceDigest,
		}
		artifacts[orderName+"/semantic-ir.json"] = struct{}{}
		artifacts[orderName+"/generated.go"] = struct{}{}
		artifacts[orderName+"/provenance.json"] = struct{}{}
		if err := writeOrderFiles(orderName, irBytes, generated, provenanceBytes); err != nil {
			return Result{}, nil, artifacts
		}
	}

	result.Orders = orders
	left := orders["A_then_B"]
	right := orders["B_then_A"]
	if left.SemanticIRDigest == right.SemanticIRDigest &&
		left.GeneratedArtifactDigest == right.GeneratedArtifactDigest &&
		left.ProvenanceDigest == right.ProvenanceDigest {
		result.Decision = Closed
		result.SemanticVerdict = "CONFLUENT"
		return result, orders, artifacts
	}

	result.Decision = Refuted
	result.SemanticVerdict = "NON_CONFLUENT"
	result.Counterexample = orderCounterexample(spec, input, left, right)
	return result, orders, artifacts
}

var activeOutputRoot string
var activeOutputRootMu sync.Mutex

func writeOrderFiles(orderName string, ir, generated, provenance []byte) error {
	if activeOutputRoot == "" {
		return errors.New("output root is not bound")
	}
	directory := filepath.Join(activeOutputRoot, orderName)
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(directory, "semantic-ir.json"), ir, 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(directory, "generated.go"), generated, 0o644); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(directory, "provenance.json"), provenance, 0o644)
}

func EvaluateToDirectory(spec Spec, input SourceInput, outputRoot string) (Result, error) {
	activeOutputRootMu.Lock()
	defer activeOutputRootMu.Unlock()
	if err := os.MkdirAll(outputRoot, 0o755); err != nil {
		return Result{}, err
	}
	activeOutputRoot = outputRoot
	defer func() { activeOutputRoot = "" }()
	result, _, _ := Evaluate(spec, input)
	if result.Schema == "" || result.Decision == "" {
		return Result{}, errors.New("semantic evaluation did not produce a terminal result")
	}
	data, err := marshalCanonical(result)
	if err != nil {
		return Result{}, err
	}
	if err := os.WriteFile(filepath.Join(outputRoot, "result.json"), data, 0o644); err != nil {
		return Result{}, err
	}
	if err := os.WriteFile(filepath.Join(outputRoot, "report.md"), []byte(markdownReport(result)), 0o644); err != nil {
		return Result{}, err
	}
	return result, nil
}

func copyState(state map[string]string) map[string]string {
	result := make(map[string]string, len(state))
	for key, value := range state {
		result[key] = value
	}
	return result
}

func sortedOperationIDs(operations map[string]Operation) []string {
	ids := make([]string, 0, len(operations))
	for id := range operations {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

func preflight(spec Spec, input SourceInput) (string, []string, string) {
	if input.SourceID == "" || input.Contract == "" || input.Toolchain == "" || input.Runner == "" {
		return "source, contract, toolchain, and runner must all be declared", []string{"input:identity"}, Unknown
	}
	if len(input.Operations) != 2 {
		return "exactly two parallel change operations are required", []string{"input:operations"}, Unknown
	}
	for _, orderName := range []string{"A_then_B", "B_then_A"} {
		if !validApplicationOrder(spec.ApplicationOrders[orderName]) {
			return "both declared application orders must apply each required operation exactly once", []string{"meta:application-orders:" + orderName}, Unknown
		}
	}
	inputOperations := operationMap(input.Operations)
	specOperations := operationMap(spec.ChangeOperations)
	blocked := make([]string, 0)
	for _, id := range []string{"change-a", "change-b"} {
		operation, ok := inputOperations[id]
		if !ok {
			blocked = append(blocked, "operation:"+id)
			continue
		}
		if _, ok := specOperations[id]; !ok {
			blocked = append(blocked, "meta-operation:"+id)
		}
		for _, field := range spec.Guardrails.RequiredOperationFields {
			if requiredOperationFieldMissing(operation, field) {
				blocked = append(blocked, "operation:"+id+":"+field)
			}
		}
		for _, dependency := range operation.DependsOn {
			if _, ok := inputOperations[dependency]; !ok {
				blocked = append(blocked, "dependency:"+dependency)
			}
		}
	}
	for _, required := range spec.Guardrails.RequiredInputPaths {
		if _, ok := input.Baseline[required]; !ok {
			blocked = append(blocked, "input:"+required)
		}
	}
	if len(blocked) > 0 {
		return "declared input does not provide the semantic frontier required for a closed decision", blocked, Unknown
	}
	for _, operation := range input.Operations {
		for _, patch := range operation.Patches {
			inScope := false
			for _, scope := range operation.Scope {
				if pathWithin(patch.Path, scope) {
					inScope = true
					break
				}
			}
			if !inScope {
				return "operation writes outside its declared semantic scope", []string{"guardrail:write-declared-scope-only", "operation:" + operation.ID + ":" + patch.Path}, Refuted
			}
			for _, prefix := range spec.Guardrails.ForbiddenPathPrefixes {
				if pathWithin(patch.Path, prefix) {
					return "guardrail forbids the requested semantic path", []string{"guardrail:" + prefix, "operation:" + operation.ID + ":" + patch.Path}, Refuted
				}
			}
		}
	}
	return "", nil, ""
}

func validApplicationOrder(order []string) bool {
	if len(order) != 2 {
		return false
	}
	return (order[0] == "change-a" && order[1] == "change-b") ||
		(order[0] == "change-b" && order[1] == "change-a")
}

func orderCounterexample(spec Spec, input SourceInput, left, right OrderEvidence) *Counterexample {
	inputOperations := operationMap(input.Operations)
	for _, operationID := range []string{"change-a", "change-b"} {
		leftOperation := inputOperations[operationID]
		for _, leftPatch := range leftOperation.Patches {
			for _, rightID := range []string{"change-a", "change-b"} {
				if operationID == rightID {
					continue
				}
				for _, rightPatch := range inputOperations[rightID].Patches {
					if leftPatch.Path == rightPatch.Path && leftPatch.Value != rightPatch.Value {
						return &Counterexample{
							Kind:     "order-dependent-path",
							Path:     leftPatch.Path,
							Reason:   "the last writer differs by application order",
							AThenB:   left.SemanticIRDigest,
							BThenA:   right.SemanticIRDigest,
							Frontier: []string{"operation:" + operationID + "->" + leftPatch.Path, "operation:" + rightID + "->" + rightPatch.Path},
						}
					}
				}
			}
		}
	}
	return &Counterexample{
		Kind:     "canonical-digest-mismatch",
		Reason:   "canonical IR, generated artifact, or provenance differs between orders",
		AThenB:   left.SemanticIRDigest,
		BThenA:   right.SemanticIRDigest,
		Frontier: []string{"order:A_then_B", "order:B_then_A"},
	}
}

func markdownReport(result Result) string {
	text := fmt.Sprintf("# Gooo semantic change confluence report\n\n- Case: `%s`\n- Decision: `%s`\n- Semantic verdict: `%s`\n- merge_authority: `%d`\n", result.CaseID, result.Decision, result.SemanticVerdict, result.MergeAuthority)
	if result.Unknown != nil {
		text += fmt.Sprintf("\n## UNKNOWN frontier\n\n- Stage: `%s`\n- Step: `%s`\n- Reason: %s\n- Unknown class: `%s`\n- Next operation: %s\n- Blocked by: `%s`\n", result.Unknown.Stage, result.Unknown.Step, result.Unknown.Reason, result.Unknown.UnknownClass, result.Unknown.NextOperation, join(result.Unknown.BlockedBy))
	}
	if result.Counterexample != nil {
		text += fmt.Sprintf("\n## Counterexample\n\n- Kind: `%s`\n- Reason: %s\n- Frontier: `%s`\n", result.Counterexample.Kind, result.Counterexample.Reason, join(result.Counterexample.Frontier))
	}
	if len(result.Orders) > 0 {
		text += "\n## Canonical order evidence\n\n| Order | Semantic IR | Generated artifact | Provenance |\n|---|---|---|---|\n"
		for _, name := range []string{"A_then_B", "B_then_A"} {
			order := result.Orders[name]
			text += fmt.Sprintf("| `%s` | `%s` | `%s` | `%s` |\n", name, order.SemanticIRDigest, order.GeneratedArtifactDigest, order.ProvenanceDigest)
		}
	}
	return text
}

func join(values []string) string {
	result := ""
	for index, value := range values {
		if index > 0 {
			result += ", "
		}
		result += "`" + value + "`"
	}
	return result
}
