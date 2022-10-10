package goldentest

import (
	"flag"
	"os"
	"path/filepath"
	"testing"
)

var update = flag.Bool("update", false, "Update the golden files instead of failing.")

func Check(t *testing.T, rerunCmd, goldenPath, actual string) {
	if *update {
		if err := os.MkdirAll(filepath.Dir(goldenPath), os.ModePerm); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(goldenPath, []byte(actual), 0666); err != nil {
			t.Fatal(err)
		}
	}

	runHelp := "Run '" + rerunCmd + " -update' to update golden file"

	want, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("Unable to read golden file. %v. Error: %v", runHelp, err)
	}

	if actual != string(want) {
		t.Errorf("Actual result didn't match golden file. %v and examine the Git diff to determine if the change is acceptable.", runHelp)
	}
}
