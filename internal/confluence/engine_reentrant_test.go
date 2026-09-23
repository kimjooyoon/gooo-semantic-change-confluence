package confluence

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestEvaluateToDirectoryIsReentrant(t *testing.T) {
	const runs = 32
	roots := make([]string, runs)
	for index := range roots {
		roots[index] = t.TempDir()
	}

	start := make(chan struct{})
	errors := make(chan error, runs)
	var group sync.WaitGroup
	for index, root := range roots {
		index, root := index, root
		group.Add(1)
		go func() {
			defer group.Done()
			<-start
			result, err := EvaluateToDirectory(testSpec(), testInput("reentrant", "receipt.currency", "KRW", "receipt.tax_code", "VAT"), root)
			if err != nil {
				errors <- err
				return
			}
			if result.Decision != Closed {
				errors <- fmt.Errorf("run %d decision = %q", index, result.Decision)
				return
			}
			for _, order := range []string{"A_then_B", "B_then_A"} {
				for _, artifact := range []string{"semantic-ir.json", "generated.go", "provenance.json"} {
					if _, err := os.Stat(filepath.Join(root, order, artifact)); err != nil {
						errors <- fmt.Errorf("run %d missing %s/%s: %w", index, order, artifact, err)
						return
					}
				}
			}
			errors <- nil
		}()
	}
	close(start)
	group.Wait()
	close(errors)
	for err := range errors {
		if err != nil {
			t.Fatal(err)
		}
	}
}
