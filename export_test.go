/*
 * Copyright © 2019 Hedzr Yeh.
 */

package cmdr

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/hedzr/cmdr/tool"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/hedzr/logex"
	"gopkg.in/hedzr/errors.v2"
)

// Worker returns unexported worker for testing
func Worker() *ExecWorker {
	return internalGetWorker()
}

// Worker2 + shouldIgnoreWrongEnumValue
func Worker2(b bool) *ExecWorker {
	internalGetWorker().shouldIgnoreWrongEnumValue = b
	return internalGetWorker()
}

// Worker3 + shouldIgnoreWrongEnumValue
func Worker3(root *RootCommand) *ExecWorker {
	w := internalGetWorker()
	w.shouldIgnoreWrongEnumValue = false
	w.rootCommand = root
	return w
}

// BacktraceCmdNames _
func BacktraceCmdNames(cmd *Command, v bool) string {
	return backtraceCmdNames(cmd, v)
}

// // ResetWorker function
// func ResetWorker() {
// 	InternalResetWorker()
// }

// InternalResetWorker is an internal helper, esp for debugging
func InternalResetWorker() (w *ExecWorker) {
	uniqueWorkerLock.Lock()
	w = internalResetWorkerNoLock()
	noResetWorker = false
	uniqueWorkerLock.Unlock()
	return
}

// InternalResetWorkerNoLock is an internal helper, esp for debugging
func InternalResetWorkerNoLock() (w *ExecWorker) {
	w = internalResetWorkerNoLock()
	return
}

// ResetRootInWorker function
func ResetRootInWorker() {
	uniqueWorkerLock.Lock()
	w := internalResetWorkerNoLock()
	w.rootCommand = nil
	uniqueWorkerLock.Unlock()
}

func TestEmptyUnknownOptionHandler(t *testing.T) {
	emptyUnknownOptionHandler(false, "", nil, nil)
}

func TestTplApply(t *testing.T) {
	tplApply("{{ .dkl }}", &struct{ sth bool }{false})
}

func tLog(a ...interface{}) {}

func TestFlag(t *testing.T) {
	ResetOptions()
	ResetRootInWorker()
	internalGetWorker().rxxtPrefixes = []string{}
	t.Log(wrapWithRxxtPrefix("x"))
	internalGetWorker().rxxtPrefixes = []string{"app"}
	InternalResetWorker()

	noResetWorker = false
	tLog(GetStringR("version"))
	noResetWorker = true
	tLog(GetStringR("version"))

	t.Log(IsDebuggerAttached())
	t.Log(InTesting())
	t.Log(InDevelopingTime())
	SetDebugMode(false)
	t.Log(GetDebugMode())
	t.Log(InDevelopingTime())
	t.Log(InDebugging())
	SetDebugMode(true)
	t.Log(GetDebugMode())
	t.Log(InDevelopingTime())
	SetTraceMode(false)
	t.Log(GetTraceMode())
	SetTraceMode(true)
	t.Log(GetTraceMode())
	t.Log(InDockerEnv())
	t.Log(tool.StripPrefix("8.yes", "8."))
	t.Log(tool.IsDigitHeavy("not-digit"))
	t.Log(tool.IsDigitHeavy("8-is-not-digit"))

	in := bytes.NewBufferString("\n")
	tool.PressEnterToContinue(in, "ok...")
	in = bytes.NewBufferString("\n")
	tool.PressEnterToContinue(in)

	in = bytes.NewBufferString("\n")
	t.Log(tool.PressAnyKeyToContinue(in, "ok..."))
	in = bytes.NewBufferString("\n")
	t.Log(tool.PressAnyKeyToContinue(in))

	isTypeFloat(reflect.TypeOf(8).Kind())
	isTypeFloat(reflect.TypeOf(8.9).Kind())

	isTypeComplex(reflect.TypeOf(8).Kind())
	isTypeComplex(reflect.TypeOf(8.9).Kind())
	isTypeComplex(reflect.TypeOf(8.9 + 0i).Kind())
	isTypeComplex(reflect.TypeOf(8.9 - 2i).Kind())

	x := tool.SavedOsArgs
	defer func() {
		tool.SavedOsArgs = x
	}()
	tool.SavedOsArgs = []string{"xx.test"}
	t.Log(InTesting())
	tool.SavedOsArgs = []string{"xx.runtime"}
	t.Log(InTesting())
	tool.SavedOsArgs = []string{"xx.runtime", "-test.v"}
	t.Log(InTesting())

	var rootCmdX = &RootCommand{
		Command: Command{
			BaseOpt: BaseOpt{
				Name: "consul-tags",
			},
			SubCommands: []*Command{
				{
					BaseOpt: BaseOpt{
						Name: "consul-tags",
					},
				},
				{
					BaseOpt: BaseOpt{
						Name: "consul-tags",
					},
				},
			},
		},
	}
	_ = walkFromCommand(&rootCmdX.Command, 0, 0, func(cmd *Command, index, level int) (err error) {
		if index > 0 {
			return ErrBadArg
		}
		return nil
	})
}

func dumpStacks() {
	fmt.Printf("=== BEGIN goroutine stack dump ===\n%s\n=== END goroutine stack dump ===\n", errors.DumpStacksAsString(true))
}

func TestHandlePanic(t *testing.T) {
	defer logex.CaptureLog(t).Release()
	if tool.SavedOsArgs == nil {
		tool.SavedOsArgs = os.Args
	}
	defer func() {
		os.Args = tool.SavedOsArgs
	}()

	ResetOptions()
	InternalResetWorker()

	onUnhandledErrorHandler1 := func(err interface{}) {
		// debug.PrintStack()
		// pprof.Lookup("goroutine").WriteTo(os.Stdout, 1)
		// dumpStacks()
		fmt.Println("some error handled: ", err)
	}

	v1, v2 := 11, 0
	// var cmd *Command
	var rootCmdX = &RootCommand{
		Command: Command{
			BaseOpt: BaseOpt{
				Name: "consul-tags",
			},
			SubCommands: []*Command{
				{
					BaseOpt: BaseOpt{
						Short: "dz", Full: "division-by-zero",
						Action: func(cmd *Command, args []string) (err error) {
							fmt.Println(v1 / v2)
							return
						},
					},
				},
				{
					BaseOpt: BaseOpt{
						Short: "pa", Full: "panic",
						Action: func(cmd *Command, args []string) (err error) {
							panic(8.1)
							// return
						},
					},
				},
			},
		},
	}

	// cmd = &rootCmdX.Command
	var commands = []string{
		"consul-tags dz",
		"consul-tags pa",
	}
	for _, cc := range commands {
		os.Args = strings.Split(cc, " ")
		SetInternalOutputStreams(nil, nil)
		ResetOptions()
		if err := Exec(rootCmdX,
			WithUnhandledErrorHandler(onUnhandledErrorHandler1),
			WithOnSwitchCharHit(func(parsed *Command, switchChar string, args []string) (err error) {
				return
			}),
		); err == nil {
			t.Error("BAD !! / ERROR !! / expecting an error returned without unexpected program terminated") // hi, here is not real error occurs
		}
	}

	t.Log(GetPredefinedLocations())
}

func TestNewOptions(t *testing.T) {
	newOptions()
	newOptionsWith(nil)
}

func TestUnknownXXX(t *testing.T) {
	defer logex.CaptureLog(t).Release()

	// // RaiseInterrupt(t, 16)
	// go func() {
	// 	time.Sleep(16 * time.Second)
	// 	SignalTermSignal()
	// }()

	if tool.SavedOsArgs == nil {
		tool.SavedOsArgs = os.Args
	}
	defer func() {
		os.Args = tool.SavedOsArgs
	}()

	var pkg *ptpkg
	var cmd *Command
	var args []string

	var rootCmdX = &RootCommand{
		Command: Command{
			BaseOpt: BaseOpt{
				Name: "consul-tags",
			},
		},
	}
	cmd = &rootCmdX.Command
	var commands = []string{
		"consul-tags --help -q",
	}
	for _, cc := range commands {
		os.Args = strings.Split(cc, " ")
		SetInternalOutputStreams(nil, nil)
		ResetOptions()
		if err := Exec(rootCmdX); err != nil {
			t.Fatal(err)
		}
	}

	pkg = &ptpkg{}
	unknownCommand(pkg, cmd, args)
	unknownFlagDetector(pkg, cmd, args)
}

// TestSliceConverters functions
func TestSliceConverters(t *testing.T) {
	stringSliceToInt64Slice([]string{"x"})
	intSliceToUint64Slice([]int{1})
	int64SliceToUint64Slice([]int64{1})
	uint64SliceToInt64Slice([]uint64{1})

	w := internalGetWorker()

	Set("x", []string{"1"})
	GetIntSliceR("x")
	w.rxxtOptions.GetInt64Slice("app.x")
	w.rxxtOptions.GetUint64Slice("app.x")
	Set("x", "1,2,3")
	GetIntSliceR("x")
	w.rxxtOptions.GetInt64Slice("app.x")
	w.rxxtOptions.GetUint64Slice("app.x")
	Set("x", []int{1, 2})
	GetIntSliceR("x")
	w.rxxtOptions.GetInt64Slice("app.x")
	w.rxxtOptions.GetUint64Slice("app.x")
	Set("x", []int64{1, 2})
	GetIntSliceR("x")
	w.rxxtOptions.GetInt64Slice("app.x")
	w.rxxtOptions.GetUint64Slice("app.x")
	Set("x", []uint64{1, 2})
	GetIntSliceR("x")
	w.rxxtOptions.GetInt64Slice("app.x")
	w.rxxtOptions.GetUint64Slice("app.x")
	Set("x", []byte{1, 2})
	GetIntSliceR("x")
	w.rxxtOptions.GetInt64Slice("app.x")
	w.rxxtOptions.GetUint64Slice("app.x")
	Set("x", 57)
	GetIntSliceR("x")
	w.rxxtOptions.GetInt64Slice("app.x")
	w.rxxtOptions.GetUint64Slice("app.x")

	mxIx("", "")
}

func (pkg *ptpkg) setOwner(cmd *Command) {
	if pkg.flg != nil {
		pkg.flg.owner = cmd
	}
}

// TestPtpkgToggleGroup functions
func TestPtpkgToggleGroup(t *testing.T) {
	pkg := &ptpkg{flg: &Flag{
		ToggleGroup: "XX",
	}}
	pkg.setOwner(&Command{
		Flags: []*Flag{
			{
				ToggleGroup: "XX",
			},
			{
				ToggleGroup: "XX",
			},
		},
	})

	pkg.tryToggleGroup()

	pkg = &ptpkg{flg: &Flag{
		DefaultValue: time.Second,
	}}
	_ = pkg.tryExtractingOthers([]string{}, reflect.Chan)
	_ = pkg.tryExtractingOthers([]string{"sss"}, reflect.Int)
	_ = pkg.processExternalTool()
}

// ExecWith is main entry of `cmdr`.
// for testing
func ExecWith(rootCmd *RootCommand, beforeXrefBuildingX, afterXrefBuiltX HookFunc) (err error) {
	w := internalGetWorker()

	if beforeXrefBuildingX != nil {
		w.beforeXrefBuilding = append(w.beforeXrefBuilding, beforeXrefBuildingX)
	}
	if afterXrefBuiltX != nil {
		w.afterXrefBuilt = append(w.afterXrefBuilt, afterXrefBuiltX)
	}

	_, err = w.InternalExecFor(rootCmd, os.Args)
	return
}

// SetInternalOutputStreams sets the internal output streams for debugging
// for testing
func SetInternalOutputStreams(out, err *bufio.Writer) {
	w := internalGetWorker()
	w.defaultStdout = out
	w.defaultStderr = err

	if w.defaultStdout == nil {
		w.defaultStdout = bufio.NewWriterSize(os.Stdout, 16384)
	}
	if w.defaultStderr == nil {
		w.defaultStderr = bufio.NewWriterSize(os.Stderr, 16384)
	}
}

// SetPredefinedLocationsForTesting
// for testing
func SetPredefinedLocationsForTesting(locations ...string) {
	internalGetWorker().predefinedLocations = locations
}

func TestNewError(t *testing.T) {

	errWrongEnumValue := newErrTmpl("unexpected enumerable value '%s' for option '%s', under command '%s'")

	err := newError(false, errWrongEnumValue, "ds", "head", "server")
	println(err)

	err = newError(true, newErr("unexpected enumerable value"))
	println(err.Error())

	err = newErrorWithMsg("Holo", errors.New("unexpected enumerable value"))
	println(err.Error())

	var perr *os.PathError
	err = newErrorWithMsg("hooloo", &os.PathError{Err: io.EOF, Op: "find", Path: "/"})
	if errors.As(err, &perr) {
		t.Logf("As() ok: %+v", *perr)
	} else {
		t.Fatal("As() failed: expect it is a os.PathError{}")
	}

	if !err.(*errors.WithStackInfo).As(&perr) {
		t.Fatal("As() failed: expect it is a os.PathError{}")
	}

	if !err.(*errors.WithStackInfo).Is(perr) {
		t.Fatal("As() failed: expect it is a os.PathError{}")
	}

	// errWrongEnumValue = newErrTmpl("unexpected enumerable value '%s' for option '%s', under command '%s'")
	// _ = errWrongEnumValue.Template("x").Format().Msg("x %v", 1).Nest(err)
}
