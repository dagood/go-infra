// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package submodule

import (
	"reflect"
	"testing"
)

func TestResetUpdateCmdReference(t *testing.T) {
	cmd := resetUpdateCmd("/repo", "/reference/repo", true)
	want := []string{"git", "submodule", "update", "--init", "--reference", "/reference/repo", "-f"}
	if !reflect.DeepEqual(cmd.Args, want) {
		t.Errorf("resetUpdateCmd args: got %q, want %q", cmd.Args, want)
	}
}
