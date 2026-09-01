package confluence

import "encoding/json"

const (
	Closed  = "CLOSED"
	Unknown = "UNKNOWN"
	Refuted = "REFUTED"
)

type Spec struct {
	Schema            string            `json:"schema"`
	Language          string            `json:"language"`
	SemanticGraph     SemanticGraph     `json:"semantic_graph"`
	ChangeOperations  []Operation       `json:"change_operations"`
	ApplicationOrders map[string][]string `json:"application_orders"`
	Authorities       Authorities       `json:"authorities"`
	Guardrails        Guardrails        `json:"guardrails"`
	CanonicalCases    []CanonicalCase   `json:"canonical_cases"`
	ReductionOrder    []string          `json:"reduction_order"`
}

type SemanticGraph struct {
	Nodes []GraphNode `json:"nodes"`
	Edges []GraphEdge `json:"edges"`
}

type GraphNode struct {
	ID   string `json:"id"`
	Kind string `json:"kind"`
}

type GraphEdge struct {
	From string `json:"from"`
	To   string `json:"to"`
	Kind string `json:"kind"`
}

type Operation struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Kind       string   `json:"kind"`
	DependsOn  []string `json:"depends_on"`
	Authority  string   `json:"authority"`
	Guardrails []string `json:"guardrails"`
	Scope      []string `json:"scope"`
	Patches    []Patch  `json:"patches"`
}

type Patch struct {
	Path  string `json:"path"`
	Value string `json:"value"`
}

type Authorities struct {
	MergeAuthority int      `json:"merge_authority"`
	RuntimeWrite   int      `json:"runtime_write"`
	InputOwner     string   `json:"input_owner"`
	ConflictPairs  [][]string `json:"conflict_pairs"`
}

type Guardrails struct {
	ForbiddenPathPrefixes []string `json:"forbidden_path_prefixes"`
	RequiredInputPaths     []string `json:"required_input_paths"`
	RequiredOperationFields []string `json:"required_operation_fields"`
}

type CanonicalCase struct {
	ID          string `json:"id"`
	Kind        string `json:"kind"`
	Expected    string `json:"expected"`
	Description string `json:"description"`
}

type SourceInput struct {
	Schema     string            `json:"schema"`
	CaseID     string            `json:"case_id"`
	SourceID   string            `json:"source_id"`
	Contract   string            `json:"contract"`
	Toolchain  string            `json:"toolchain"`
	Runner     string            `json:"runner"`
	Baseline   map[string]string `json:"baseline"`
	Operations []Operation       `json:"operations"`
}

type UnknownEvidence struct {
	Stage         string   `json:"stage"`
	Step          string   `json:"step"`
	Reason        string   `json:"reason"`
	UnknownClass  string   `json:"unknown_class"`
	NextOperation string   `json:"next_operation"`
	BlockedBy     []string `json:"blocked_by"`
}

type Counterexample struct {
	Kind        string   `json:"kind"`
	Path        string   `json:"path,omitempty"`
	Reason      string   `json:"reason"`
	AThenB      string   `json:"a_then_b,omitempty"`
	BThenA      string   `json:"b_then_a,omitempty"`
	Frontier    []string `json:"frontier"`
	Authorities []string `json:"authorities,omitempty"`
}

type SemanticIR struct {
	Schema          string            `json:"schema"`
	SourceID        string            `json:"source_id"`
	CaseID          string            `json:"case_id"`
	Contract        string            `json:"contract"`
	Toolchain       string            `json:"toolchain"`
	Runner          string            `json:"runner"`
	Graph           SemanticGraph     `json:"graph"`
	State           map[string]string `json:"state"`
	Applied         []string          `json:"applied_operations"`
}

type ProvenanceEntry struct {
	Operation string `json:"operation"`
	Path      string `json:"path"`
	Value     string `json:"value"`
	Authority string `json:"authority"`
}

type OrderEvidence struct {
	ExecutionOrder       []string `json:"execution_order"`
	SemanticIRFile       string   `json:"semantic_ir_file"`
	SemanticIRDigest     string   `json:"semantic_ir_digest"`
	GeneratedArtifactFile string   `json:"generated_artifact_file"`
	GeneratedArtifactDigest string `json:"generated_artifact_digest"`
	ProvenanceFile       string   `json:"provenance_file"`
	ProvenanceDigest     string   `json:"provenance_digest"`
}

type Result struct {
	Schema          string                    `json:"schema"`
	CaseID          string                    `json:"case_id"`
	Decision        string                    `json:"decision"`
	SemanticVerdict string                    `json:"semantic_verdict"`
	MergeAuthority  int                       `json:"merge_authority"`
	Orders          map[string]OrderEvidence  `json:"orders,omitempty"`
	Unknown         *UnknownEvidence          `json:"unknown,omitempty"`
	Counterexample  *Counterexample           `json:"counterexample,omitempty"`
}

func decode(data []byte, out any) error {
	return json.Unmarshal(data, out)
}
