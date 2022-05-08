/*
 * Copyright © 2019 Hedzr Yeh.
 */

package main

import (
	"fmt"

	"github.com/hedzr/cmdr"
	"github.com/hedzr/cmdr/plugin/pprof"
)

func main() {
	// logrus.SetLevel(logrus.DebugLevel)
	// logrus.SetFormatter(&logrus.TextFormatter{ForceColors: true,})

	if err := cmdr.Exec(rootCmd,
		// To disable internal commands and flags, uncomment the following codes
		cmdr.WithBuiltinCommands(true, false, true, true, true),
		// daemon.WithDaemon(svr.NewDaemon(), nil, nil, nil),
		pprof.GetCmdrProfilingOptions("cpu"),
		// cmdr.WithHelpTabStop(40),
	); err != nil {
		fmt.Printf("Error: %v", err)
	}
}

var (
	rootCmd = &cmdr.RootCommand{
		Command: cmdr.Command{
			BaseOpt: cmdr.BaseOpt{
				Name: "short",
			},
			Flags: []*cmdr.Flag{},
			SubCommands: []*cmdr.Command{
				serverCommands,
				// msCommands,
			},
		},

		AppName:    "short",
		Version:    cmdr.Version,
		VersionInt: cmdr.VersionInt,
		Copyright:  "austr is an effective devops tool",
		Author:     "Hedzr Yeh <hedzrz@gmail.com>",
	}

	serverCommands = &cmdr.Command{
		BaseOpt: cmdr.BaseOpt{
			Short:       "s",
			Full:        "server",
			Aliases:     []string{"serve", "svr"},
			Description: "server ops: for linux service/daemon.",
		},
		Flags: []*cmdr.Flag{
			{
				BaseOpt: cmdr.BaseOpt{
					Short:       "f",
					Full:        "foreground",
					Aliases:     []string{"fg"},
					Description: "running at foreground",
				},
			},
		},
		SubCommands: []*cmdr.Command{
			{
				BaseOpt: cmdr.BaseOpt{
					Short:       "s",
					Full:        "start",
					Aliases:     []string{"run", "startup"},
					Description: "startup this system service/daemon.",
					Action: func(cmd *cmdr.Command, args []string) (err error) {
						return
					},
				},
			},
			{
				BaseOpt: cmdr.BaseOpt{
					Short:       "t",
					Full:        "stop",
					Aliases:     []string{"stp", "halt", "pause"},
					Description: "stop this system service/daemon.",
				},
			},
			{
				BaseOpt: cmdr.BaseOpt{
					Short:       "r",
					Full:        "restart",
					Aliases:     []string{"reload"},
					Description: "restart this system service/daemon.",
				},
			},
			{
				BaseOpt: cmdr.BaseOpt{
					Full:        "status",
					Aliases:     []string{"st"},
					Description: "display its running status as a system service/daemon.",
				},
			},
			{
				BaseOpt: cmdr.BaseOpt{
					Short:       "i",
					Full:        "install",
					Aliases:     []string{"setup"},
					Description: "install as a system service/daemon.",
				},
			},
			{
				BaseOpt: cmdr.BaseOpt{
					Short:       "u",
					Full:        "uninstall",
					Aliases:     []string{"remove"},
					Description: "remove from a system service/daemon.",
				},
			},
		},
	}
)
