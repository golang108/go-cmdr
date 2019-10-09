/*
 * Copyright © 2019 Hedzr Yeh.
 */

package daemon

import (
	"github.com/hedzr/cmdr"
	"github.com/hedzr/cmdr/plugin/daemon/impl"
	"log"

	"os"
)

// WithDaemon enables daemon plugin:
// - add daemon commands and sub-commands: start/run, stop, restart/reload, status, install/uninstall
// - pidfile
// -
func WithDaemon(daemonImplX Daemon,
	modifier func(daemonServerCommand *cmdr.Command) *cmdr.Command,
	preAction func(cmd *cmdr.Command, args []string) (err error),
	postAction func(cmd *cmdr.Command, args []string),
) cmdr.ExecOption {
	return func(w *cmdr.ExecWorker) {
		daemonImpl = daemonImplX

		w.AddOnBeforeXrefBuilding(func(root *cmdr.RootCommand, args []string) {

			if modifier != nil {
				root.SubCommands = append(root.SubCommands, modifier(DaemonServerCommand))
			} else {
				root.SubCommands = append(root.SubCommands, DaemonServerCommand)
			}

			// prefix = strings.Join(append(cmdr.RxxtPrefix, "server"), ".")
			// prefix = "server"

			attachPreAction(root, preAction)
			attachPostAction(root, postAction)

		})
	}
}

// // Enable enables the daemon plugin:
// // - add daemon commands and sub-commands: start/run, stop, restart/reload, status, install/uninstall
// // - pidfile
// // - go-daemon supports
// // -
// //
// // Deprecated: from v1.5.0, replaced with WithDaemon()
// func Enable(daemonImplX Daemon,
// 	modifier func(daemonServerCommand *cmdr.Command) *cmdr.Command,
// 	preAction func(cmd *cmdr.Command, args []string) (err error),
// 	postAction func(cmd *cmdr.Command, args []string),
// ) {
// 	daemonImpl = daemonImplX
//
// 	cmdr.AddOnBeforeXrefBuilding(func(root *cmdr.RootCommand, args []string) {
//
// 		if modifier != nil {
// 			root.SubCommands = append(root.SubCommands, modifier(DaemonServerCommand))
// 		} else {
// 			root.SubCommands = append(root.SubCommands, DaemonServerCommand)
// 		}
//
// 		// prefix = strings.Join(append(cmdr.RxxtPrefix, "server"), ".")
//
// 		attachPreAction(root, preAction)
// 		attachPostAction(root, postAction)
//
// 	})
// }

func attachPostAction(root *cmdr.RootCommand, postAction func(cmd *cmdr.Command, args []string)) {
	if root.PostAction != nil {
		savedPostAction := root.PostAction
		root.PostAction = func(cmd *cmdr.Command, args []string) {
			if postAction != nil {
				postAction(cmd, args)
			}
			pidfile.Destroy()
			savedPostAction(cmd, args)
			return
		}
	} else {
		root.PostAction = func(cmd *cmdr.Command, args []string) {
			if postAction != nil {
				postAction(cmd, args)
			}
			pidfile.Destroy()
			return
		}
	}
}

func attachPreAction(root *cmdr.RootCommand, preAction func(cmd *cmdr.Command, args []string) (err error)) {
	if root.PreAction != nil {
		savedPreAction := root.PreAction
		root.PreAction = func(cmd *cmdr.Command, args []string) (err error) {
			pidfile.Create(cmd)
			logger.Setup(cmd)
			if err = savedPreAction(cmd, args); err != nil {
				return
			}
			if preAction != nil {
				err = preAction(cmd, args)
			}
			return
		}
	} else {
		root.PreAction = func(cmd *cmdr.Command, args []string) (err error) {
			pidfile.Create(cmd)
			logger.Setup(cmd)
			if preAction != nil {
				err = preAction(cmd, args)
			}
			return
		}
	}
}

func daemonStart(cmd *cmdr.Command, args []string) (err error) {
	ctx := impl.GetContext(cmd, args)
	if cmdr.GetBoolRP(keyPrefix, "foreground") {
		err = run(ctx, cmd, args)
	} else if cmd.GetHitStr() == "run" {
		err = run(ctx, cmd, args)
	} else {
		err = runAsDaemon(cmd, args)
	}
	return
}

func runAsDaemon(cmd *cmdr.Command, args []string) (err error) {
	ctx := impl.GetContext(cmd, args)
	if err := impl.Demonize(ctx); err != nil {
		log.Printf("Unable to create child process: %+v", err)
		os.Exit(impl.ErrnoForkAndDaemonFailed)
	}

	// cxt := getContext(cmd, args)
	// child, err = cxt.Reborn()
	// if err != nil {
	// 	log.Fatalln(err)
	// }
	// if child != nil {
	// 	fmt.Println("Daemon started. parent stopped.")
	// 	return
	// }
	//
	// cmdr.Set(DaemonizedKey, true)
	//
	// defer cxt.Release()

	err = run(ctx, cmd, args)
	return
}

// IsRunningInDemonizedMode returns true if you are running under demonized mode.
// false means that you're running in normal console/tty mode.
func IsRunningInDemonizedMode() bool {
	// return cmdr.GetBoolR(DaemonizedKey)
	return impl.IsRunningInDemonizedMode()
}

// SetTermSignals allows an functor to provide a list of Signals
func SetTermSignals(sig func() []os.Signal) {
	// onSetTermHandler = sig
	impl.SetTermSignals(sig)
}

// SetSigEmtSignals allows an functor to provide a list of Signals
func SetSigEmtSignals(sig func() []os.Signal) {
	// onSetSigEmtHandler = sig
	impl.SetSigEmtSignals(sig)
}

// SetReloadSignals allows an functor to provide a list of Signals
func SetReloadSignals(sig func() []os.Signal) {
	// onSetReloadHandler = sig
	impl.SetReloadSignals(sig)
}

// func setupSignals() {
// 	// for i := 1; i < 34; i++ {
// 	// 	daemon.SetSigHandler(termHandler, syscall.Signal(i))
// 	// }
//
// 	signals := []os.Signal{syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGABRT, syscall.SIGINT, syscall.SIGKILL, syscall.SIGUSR1, syscall.SIGUSR2}
// 	if onSetTermHandler != nil {
// 		signals = onSetTermHandler()
// 	}
// 	daemon.SetSigHandler(termHandler, signals...)
//
// 	signals = []os.Signal{syscall.Signal(0x7)}
// 	if onSetSigEmtHandler != nil {
// 		signals = onSetSigEmtHandler()
// 	}
// 	daemon.SetSigHandler(sigEmtHandler, signals...)
//
// 	signals = []os.Signal{syscall.SIGHUP}
// 	if onSetReloadHandler != nil {
// 		signals = onSetReloadHandler()
// 	}
// 	daemon.SetSigHandler(reloadHandler, signals...)
// }

func run(ctx *impl.Context, cmd *cmdr.Command, args []string) (err error) {
	// setupSignals()
	//
	// if daemonImpl != nil {
	// 	if err = daemonImpl.OnRun(cmd, args, stop, done); err != nil {
	// 		return
	// 	}
	// }
	//
	// log.Printf("daemon ServeSignals, pid = %v", os.Getpid())
	// if err = daemon.ServeSignals(); err != nil {
	// 	log.Println("Error:", err)
	// }
	//
	// if daemonImpl != nil {
	// 	err = daemonImpl.OnStop(cmd, args)
	// }
	//
	// if err != nil {
	// 	log.Fatal("daemon terminated.", err)
	// }
	// log.Println("daemon terminated.")

	if daemonImpl != nil {
		stop, done := impl.GetChs()
		if err = daemonImpl.OnRun(cmd, args, stop, done); err != nil {
			return
		}
	}

	log.Printf("daemon ServeSignals, pid = %v", os.Getpid())
	if err = impl.ServeSignals(ctx); err != nil {
		log.Println("Error:", err)
	}

	if daemonImpl != nil {
		err = daemonImpl.OnStop(cmd, args)
	}

	if err != nil {
		log.Fatal("daemon terminated.", err)
	}
	log.Println("daemon terminated.")
	return
}

func daemonStop(cmd *cmdr.Command, args []string) (err error) {
	// getContext(cmd, args)
	//
	// p, err := daemonCtx.Search()
	// if err != nil {
	// 	fmt.Printf("%v is stopped.\n", cmd.GetRoot().AppName)
	// 	return
	// }
	//
	// if err = p.Signal(syscall.SIGTERM); err != nil {
	// 	return
	// }

	ctx := impl.GetContext(cmd, args)
	impl.Stop(cmd.GetRoot().AppName, ctx)
	return
}

func daemonRestart(cmd *cmdr.Command, args []string) (err error) {
	// getContext(cmd, args)
	//
	// p, err := daemonCtx.Search()
	// if err != nil {
	// 	fmt.Printf("%v is stopped.\n", cmd.GetRoot().AppName)
	// } else {
	// 	if err = p.Signal(syscall.SIGHUP); err != nil {
	// 		return
	// 	}
	// }

	ctx := impl.GetContext(cmd, args)
	impl.Reload(cmd.GetRoot().AppName, ctx)
	return
}

func daemonStatus(cmd *cmdr.Command, args []string) (err error) {
	// getContext(cmd, args)
	//
	// p, err := daemonCtx.Search()
	// if err != nil {
	// 	fmt.Printf("%v is stopped.\n", cmd.GetRoot().AppName)
	// } else {
	// 	fmt.Printf("%v is running as %v.\n", cmd.GetRoot().AppName, p.Pid)
	// }
	//
	// if daemonImpl != nil {
	// 	err = daemonImpl.OnStatus(&Context{Context: *daemonCtx}, cmd, p)
	// }

	ctx := impl.GetContext(cmd, args)
	present, process := impl.FindDaemonProcess(ctx)
	if present && daemonImpl != nil {
		err = daemonImpl.OnStatus(ctx, cmd, process)
	}
	return
}

func daemonInstall(cmd *cmdr.Command, args []string) (err error) {
	ctx := impl.GetContext(cmd, args)

	err = runInstaller(cmd, args)
	if err != nil {
		return
	}
	if daemonImpl != nil {
		err = daemonImpl.OnInstall(ctx /*&Context{Context: *daemonCtx}*/, cmd, args)
	}
	return
}

func daemonUninstall(cmd *cmdr.Command, args []string) (err error) {
	ctx := impl.GetContext(cmd, args)

	err = runUninstaller(cmd, args)
	if err != nil {
		return
	}
	if daemonImpl != nil {
		err = daemonImpl.OnUninstall(ctx /*&Context{Context: *daemonCtx}*/, cmd, args)
	}
	return
}

// // DaemonizedKey is the keyPath to ensure you are running in demonized mode.
// const DaemonizedKey = "demonized"

// var child *os.Process

// var onSetTermHandler func() []os.Signal
// var onSetSigEmtHandler func() []os.Signal
// var onSetReloadHandler func() []os.Signal

// var prefix string
const keyPrefix = "server"
