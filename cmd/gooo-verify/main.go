package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/gooo-semantic-change-confluence/internal/confluence"
)

func main() {
	resultPath := flag.String("result", "", "result.json emitted by gooo-confluence")
	root := flag.String("root", "", "run output directory")
	flag.Parse()
	if *resultPath == "" || *root == "" {
		fail("-result and -root are required")
	}
	data, err := os.ReadFile(*resultPath)
	if err != nil {
		fail(err.Error())
	}
	var result confluence.Result
	if err := json.Unmarshal(data, &result); err != nil {
		fail(err.Error())
	}
	if result.Schema != "gooo/semantic-confluence-result/v1" {
		fail("unexpected result schema")
	}
	switch result.Decision {
	case confluence.Unknown:
		if result.Unknown == nil || result.Unknown.Stage == "" || result.Unknown.Step == "" || result.Unknown.Reason == "" || result.Unknown.UnknownClass == "" || result.Unknown.NextOperation == "" || len(result.Unknown.BlockedBy) == 0 {
			fail("UNKNOWN result does not carry the required six-field frontier")
		}
	case confluence.Closed:
		if result.SemanticVerdict != "CONFLUENT" || len(result.Orders) != 2 {
			fail("CLOSED result is missing both order records")
		}
		for _, name := range []string{"A_then_B", "B_then_A"} {
			order, ok := result.Orders[name]
			if !ok {
				fail("missing order " + name)
			}
			verifyFile(*root, order.SemanticIRFile, order.SemanticIRDigest)
			verifyFile(*root, order.GeneratedArtifactFile, order.GeneratedArtifactDigest)
			verifyFile(*root, order.ProvenanceFile, order.ProvenanceDigest)
		}
	case confluence.Refuted:
		if result.SemanticVerdict != "NON_CONFLUENT" || result.Counterexample == nil || len(result.Counterexample.Frontier) == 0 || result.Counterexample.Reason == "" {
			fail("REFUTED result is missing its minimal counterexample")
		}
	default:
		fail("unknown top-level decision; fail closed")
	}
	fmt.Printf("verified case=%s decision=%s\n", result.CaseID, result.Decision)
}

func verifyFile(root, relative, expected string) {
	data, err := os.ReadFile(filepath.Join(root, relative))
	if err != nil {
		fail(err.Error())
	}
	sum := sha256.Sum256(data)
	actual := "sha256:" + hex.EncodeToString(sum[:])
	if actual != expected {
		fail(fmt.Sprintf("digest mismatch for %s: got %s want %s", relative, actual, expected))
	}
}

func fail(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}
