package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/kimjooyoon/gooo-semantic-change-confluence/internal/confluence"
)

func main() {
	specPath := flag.String("spec", ".gooo/confluence.gooo", "Gooo semantic rules")
	inputPath := flag.String("input", "", "source .gooo input")
	outputPath := flag.String("output", ".ci/run", "caller-owned output directory")
	flag.Parse()
	if *inputPath == "" {
		fail("-input is required")
	}
	spec, err := confluence.LoadSpec(*specPath)
	if err != nil {
		fail(err.Error())
	}
	input, err := confluence.LoadInput(*inputPath)
	if err != nil {
		fail(err.Error())
	}
	result, err := confluence.EvaluateToDirectory(spec, input, *outputPath)
	if err != nil {
		fail(err.Error())
	}
	fmt.Printf("case=%s decision=%s semantic_verdict=%s output=%s\n", result.CaseID, result.Decision, result.SemanticVerdict, *outputPath)
	if result.Decision == confluence.Unknown {
		return
	}
}

func fail(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}
