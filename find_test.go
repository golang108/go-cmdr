// Copyright © 2020 Hedzr Yeh.

package cmdr_test

import (
	"github.com/hedzr/cmdr"
	"os"
	"strings"
	"testing"
)

func TestFinds(t *testing.T) {
	t.Log("finds")
	cmdr.InternalResetWorkerForTest()
	cmdr.ResetOptions()

	cmdr.Set("no-watch-conf-dir", true)

	// copyRootCmd = rootCmdForTesting
	var rootCmdX = &cmdr.RootCommand{
		Command: cmdr.Command{
			BaseOpt: cmdr.BaseOpt{
				Name: "consul-tags",
			},
		},
	}
	t.Log("rootCmdForTesting", rootCmdX)

	var commands = []string{
		"consul-tags --help -q",
	}
	for _, cc := range commands {
		os.Args = strings.Split(cc, " ")
		cmdr.SetInternalOutputStreams(nil, nil)

		if err := cmdr.Exec(rootCmdX); err != nil {
			t.Fatal(err)
		}
	}

	if cmdr.InTesting() {
		cmdr.FindSubCommand("generate", nil)
		cmdr.FindFlag("generate", nil)
		cmdr.FindSubCommandRecursive("generate", nil)
		cmdr.FindFlagRecursive("generate", nil)
	} else {
		t.Log("noted")
	}

	cmdr.GetRemainArgs()
	cmdr.EnableShellPager(true)
	cmdr.EnableShellPager(false)
	resetOsArgs()
}
