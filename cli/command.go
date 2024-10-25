package cli

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync/atomic"

	"github.com/hedzr/is"
	"github.com/hedzr/is/states"

	"github.com/hedzr/cmdr/v2/internal/tool"
	"github.com/hedzr/cmdr/v2/pkg/dir"
	"github.com/hedzr/cmdr/v2/pkg/exec"
	"github.com/hedzr/cmdr/v2/pkg/logz"

	"gopkg.in/hedzr/errors.v3"
)

//
//
//

func (c *CmdS) IsRoot() bool {
	if x, ok := c.root.Cmd.(*CmdS); ok && x == c {
		return c.owner == nil
	}
	return false
	// return c.root.Cmd == c && c.owner == nil
}

func (c *CmdS) HasFlag(longTitle string) (f *Flag, ok bool) {
	f, ok = c.longFlags[longTitle]
	return
}

// func (c *CmdS) Root() *RootCommand      { return c.root }
// func (c *CmdS) Owner() *CmdS         { return c.owner }

func (c *CmdS) SubCommands() []*CmdS { return c.commands }
func (c *CmdS) Flags() []*Flag       { return c.flags }
func (c *CmdS) String() string {
	var sb strings.Builder
	_, _ = sb.WriteString("Cmd{'")
	// _, _ = sb.WriteString(c.GetTitleName())
	_, _ = sb.WriteString(c.GetDottedPath())
	_, _ = sb.WriteString("'}")
	return sb.String()
}

// TailPlaceHolder is a string at end of usage line in help screen.
//
// In help screen, a command's usage line generally has the following form:
//
//	<app-name> <sub-commands> [<options>] [<positional-args>...]
//
// The text of <positional-args> is exact TailPlaceHolder. Set TailPlaceHolder
// to "files..." might be meaningful for "load" command, looks like:
//
//	<app> yaml-formatter load [<options>] files...
func (c *CmdS) TailPlaceHolder() string { return strings.Join(c.tailPlaceHolders, " ") }

// RedirectTo provides the real command target for current CmdS.
//
// Suppose command [app build] is being redirected to [app gcc build]. There
// [app build] is a shortcut to its full commands [app gcc build].
func (c *CmdS) RedirectTo() (dottedPath string) { return c.redirectTo }

// GetQuotedGroupName returns the group name quoted string.
func (c *CmdS) GetQuotedGroupName() string {
	if strings.TrimSpace(c.group) == "" {
		return ""
	}
	i := strings.Index(c.group, ".")
	if i >= 0 {
		return fmt.Sprintf("[%v]", c.group[i+1:])
	}
	return fmt.Sprintf("[%v]", c.group)
}

// GetExpandableNamesArray returns the names array of command, includes short name and long name.
func (c *CmdS) GetExpandableNamesArray() []string {
	var a []string
	if len(c.Long) > 0 {
		a = append(a, c.Long)
	}
	if len(c.Short) > 0 {
		a = append(a, c.Short)
	}
	return a
}

// GetExpandableNames returns the names comma splitted string.
func (c *CmdS) GetExpandableNames() string {
	a := c.GetExpandableNamesArray()
	if len(a) == 1 {
		return a[0]
	} else if len(a) > 1 {
		return fmt.Sprintf("{%v}", strings.Join(a, ","))
	}
	return c.name
}

//
//

func (c *CmdS) AppendTailPlaceHolder(placeHolder ...string) {
	c.tailPlaceHolders = append(c.tailPlaceHolders, placeHolder...)
}

func (c *CmdS) SetTailPlaceHolder(placeHolders ...string) { c.tailPlaceHolders = placeHolders }
func (c *CmdS) SetRedirectTo(dottedPath string)           { c.redirectTo = dottedPath }
func (c *CmdS) SetPresetCmdLines(args ...string)          { c.presetCmdLines = args }
func (c *CmdS) SetInvokeProc(str string)                  { c.invokeProc = str }
func (c *CmdS) SetInvokeShell(str string)                 { c.invokeShell = str }
func (c *CmdS) SetShell(str string)                       { c.shell = str }

//
//

func (c *CmdS) AddSubCommand(child *CmdS, callbacks ...func(cc *CmdS)) { //nolint:revive
	if child == nil {
		return
	}

	for _, cc := range c.commands {
		if cc == child || cc.EqualTo(child) {
			return
		}
	}
	for _, cb := range callbacks {
		if cb != nil {
			cb(child)
		}
	}
	c.commands = append(c.commands, child)
	child.owner = c
	child.root = c.root
}

func (c *CmdS) AddFlag(child *Flag, callbacks ...func(ff *Flag)) { //nolint:revive
	if child == nil {
		return
	}

	for _, cc := range c.flags {
		if cc == child || cc.EqualTo(child) {
			return
		}
	}
	for _, cb := range callbacks {
		if cb != nil {
			cb(child)
		}
	}
	c.flags = append(c.flags, child)
	child.owner = c
	child.root = c.root
}

//
//
//

// SetOnMatched adds the onMatched handler to a command
func (c *CmdS) SetOnMatched(functions ...OnCommandMatchedHandler) {
	c.onMatched = append(c.onMatched, functions...)
}

func (c *CmdS) SetOnEvaluateSubCommands(handler OnEvaluateSubCommands) {
	c.onEvalSubcommands = &struct{ cb OnEvaluateSubCommands }{cb: handler}
}

func (c *CmdS) SetOnEvaluateSubCommandsOnce(handler OnEvaluateSubCommands) {
	c.onEvalSubcommandsOnce = &struct {
		cb       OnEvaluateSubCommands
		invoked  bool
		commands []Cmd
	}{cb: handler}
}

func (c *CmdS) SetOnEvaluateFlags(handler OnEvaluateFlags) {
	c.onEvalFlags = &struct{ cb OnEvaluateFlags }{cb: handler}
}

func (c *CmdS) SetOnEvaluateFlagsOnce(handler OnEvaluateFlags) {
	c.onEvalFlagsOnce = &struct {
		cb      OnEvaluateFlags
		invoked bool
		flags   []*Flag
	}{cb: handler}
}

//

func (c *CmdS) OnEvalSubcommands() OnEvaluateSubCommands {
	if c.onEvalSubcommands == nil {
		return nil
	}
	return c.onEvalSubcommands.cb
}

func (c *CmdS) OnEvalSubcommandsOnce() OnEvaluateSubCommands {
	if c.onEvalSubcommandsOnce == nil {
		return nil
	}
	return c.onEvalSubcommandsOnce.cb
}

func (c *CmdS) OnEvalSubcommandsOnceInvoked() bool {
	if c.onEvalSubcommandsOnce == nil {
		return false
	}
	return c.onEvalSubcommandsOnce.invoked
}

func (c *CmdS) OnEvalSubcommandsOnceCache() []Cmd {
	if c.onEvalSubcommandsOnce == nil {
		return nil
	}
	return c.onEvalSubcommandsOnce.commands
}

func (c *CmdS) OnEvalSubcommandsOnceSetCache(list []Cmd) {
	if c.onEvalSubcommandsOnce == nil {
		return
	}
	c.onEvalSubcommandsOnce.commands = list
	c.onEvalSubcommandsOnce.invoked = true
}

//

func (c *CmdS) OnEvalFlags() OnEvaluateFlags {
	if c.onEvalFlags == nil {
		return nil
	}
	return c.onEvalFlags.cb
}

func (c *CmdS) OnEvalFlagsOnce() OnEvaluateFlags {
	if c.onEvalFlagsOnce == nil {
		return nil
	}
	return c.onEvalFlagsOnce.cb
}

func (c *CmdS) OnEvalFlagsOnceInvoked() bool {
	if c.onEvalFlagsOnce == nil {
		return false
	}
	return c.onEvalFlagsOnce.invoked
}

func (c *CmdS) OnEvalFlagsOnceCache() []*Flag {
	if c.onEvalFlagsOnce == nil {
		return nil
	}
	return c.onEvalFlagsOnce.flags
}

func (c *CmdS) OnEvalFlagsOnceSetCache(list []*Flag) {
	if c.onEvalFlagsOnce == nil {
		return
	}
	c.onEvalFlagsOnce.flags = list
	c.onEvalFlagsOnce.invoked = true
}

//

func (c *CmdS) SetHitTitle(title string) {
	c.hitTitle = title
	c.hitTimes++
}
func (c *CmdS) HitTitle() string { return c.hitTitle }
func (c *CmdS) HitTimes() int    { return c.hitTimes }

func (c *CmdS) ShortName() string    { return c.Short }
func (c *CmdS) ShortNames() []string { return c.Shorts() }
func (c *CmdS) AliasNames() []string { return c.Aliases }

// func (c *CmdS) OwnerCmd() Cmd      { return c.owner }

func (c *CmdS) OwnerIsValid() bool { return c.OwnerIsNotNil() && c.owner != c }

func (c *CmdS) HeadLikeFlag() *Flag      { return c.headLikeFlag }
func (c *CmdS) SetHeadLikeFlag(ff *Flag) { c.headLikeFlag = ff }

//

func (c *CmdS) CanInvoke() bool { return c.onInvoke != nil }

// SetPostActions adds the post-action to a command
func (c *CmdS) SetPostActions(functions ...OnPostInvokeHandler) {
	c.postActions = append(c.postActions, functions...)
}

// SetPreActions adds the pre-action to a command
func (c *CmdS) SetPreActions(functions ...OnPreInvokeHandler) {
	c.preActions = append(c.preActions, functions...)
}

// SetAction adds the onInvoke action to a command
func (c *CmdS) SetAction(fn OnInvokeHandler) {
	c.onInvoke = fn
}

func (c *CmdS) Invoke(ctx context.Context, args []string) (err error) {
	var deferActions func(errInvoked error)
	if deferActions, err = c.RunPreActions(ctx, c, args); err != nil {
		logz.VerboseContext(ctx, "cmd.RunPreActions failed", "err", err)
		return
	}
	defer func() { deferActions(err) }() // err must be delayed caught here

	logz.VerboseContext(ctx, "cmd.Invoke()", "onInvoke", c.onInvoke)
	if c.onInvoke != nil {
		err = c.onInvoke(ctx, c, args)
	}
	return
}

func (c *CmdS) RunPreActions(ctx context.Context, cmd Cmd, args []string) (deferAction func(errInvoked error), err error) { //nolint:revive
	ec := errors.New("[PRE-INVOKE]")
	defer ec.Defer(&err)
	if c.root.Cmd != c {
		for _, a := range c.root.preActions {
			if a != nil {
				ec.Attach(a(ctx, cmd, args))
			}
		}
	}

	for _, a := range c.preActions {
		if a != nil {
			ec.Attach(a(ctx, cmd, args))
		}
	}

	if !ec.IsEmpty() {
		deferAction = func(errInvoked error) {}
		return
	}

	deferAction = c.getDeferAction(ctx, cmd, args)
	return
}

func (c *CmdS) getDeferAction(ctx context.Context, cmd Cmd, args []string) func(errInvoked error) { //nolint:revive
	return func(errInvoked error) {
		ecp := errors.New("[POST-INVOKE]")

		// if !errors.Iss(errInvoked, ErrShouldStop, ErrShouldFallback) { // condition is true if errInvoked is nil
		// 	ecp.Attach(errInvoked) // no matter, attaching a nil error is no further effect
		// }

		for _, a := range c.postActions {
			if a != nil {
				ecp.Attach(a(ctx, cmd, args, errInvoked))
			}
		}
		if c.root.Cmd != c {
			for _, a := range c.root.postActions {
				if a != nil {
					ecp.Attach(a(ctx, cmd, args, errInvoked))
				}
			}
		}

		if !ecp.IsEmpty() {
			logz.Panic("Error(s) occurred when running post-actions:", "error", ecp.Error())
		}
	}
}

// func (c *CmdS) RunPreActions(cmd *CmdS, args []string) (deferAction func(errInvoked error), err error) {
// 	var ec = errors.New("[PRE-INVOKE]")
// 	defer ec.Defer(&err)
// 	if c.preInvoke != nil {
// 		ec.Attach(c.preInvoke(cmd, args))
// 	}
//
// 	deferAction = func(errInvoked error) {
// 		var ecp = errors.New("[POST-INVOKE]")
// 		if c.postInvoke != nil {
// 			ecp.Attach(c.postInvoke(cmd, args, errInvoked))
// 		}
// 		if !ecp.IsEmpty() {
// 			logz.Fatalf("Error(s) occurred when running post-actions: %v", ecp)
// 		}
// 		return
// 	}
// 	return
// }

//
//
//

// EnsureTree associates owner and app between all subCommands and app/runner/rootCommand.
// EnsureTree links all commands as a tree (make root and owner linked).
//
// This function will be called for running time once (see also cmdr.Run()).
func (c *CmdS) EnsureTree(ctx context.Context, app App, root *RootCommand) {
	if atomic.CompareAndSwapInt32(&root.linked, 0, 1) {
		logz.VerboseContext(ctx, "cmd.EnsureTree (Once) -> linking to root and owner", "root", root)
		c.ensureTreeAlways(ctx, app, root)
	}
}

// EnsureTreeAlways associates owner and app between all subCommands and app/runner/rootCommand.
// EnsureTreeAlways links all commands as a tree (make root and owner linked).
//
// This func is called only in building command system (see also builder.postBuild).
func (c *CmdS) EnsureTreeAlways(ctx context.Context, app App, root *RootCommand) {
	logz.DebugContext(ctx, "cmd.EnsureTreeAlways -> linking to root and owner", "root", root)
	c.ensureTreeAlways(ctx, app, root)
}

func (c *CmdS) ensureTreeAlways(ctx context.Context, app App, root *RootCommand) {
	root.app = app // link RootCommand.app to app
	root.SetName(app.Name())
	c.ensureTreeR(ctx, app, root)
}

// ensureTreeR link CmdS.owner to its parent, and CmdS.root to root.
// ensureTreeR links all commands as a tree (make root and owner linked).
func (c *CmdS) ensureTreeR(ctx context.Context, app App, root *RootCommand) { //nolint:unparam,revive
	c.WalkEverything(ctx, func(cc, pp Cmd, ff *Flag, cmdIndex, flgIndex, level int) {
		if cx, ok1 := cc.(*CmdS); ok1 {
			logz.VerboseContext(ctx, "    .|. ensureTreeR, (+owner,+root), CmdS:", "cmd", cx)
			if pp != nil {
				cx.owner = pp.(*CmdS)
			} else {
				cx.owner = nil
			}
			cx.root, _ = root, app
			if ff != nil {
				ff.owner, ff.root = cx, root
			}
		} else if ff == nil {
			// others: commands
			logz.VerboseContext(ctx, "    .|. ensureTreeR, (+owner,+root), Non-CmdS:", "cmd", cc)
			if cx, ok := cc.(interface{ SetOwner(o *CmdS) }); ok {
				if pp != nil {
					cx.SetOwner(pp.(*CmdS))
				} else {
					cx.SetOwner(nil)
				}
			}
			if cx, ok := cc.(interface{ SetOwner(o Cmd) }); ok {
				cx.SetOwner(pp)
			}
			if cx, ok := cc.(interface{ SetOwnerCmd(o Cmd) }); ok {
				cx.SetOwnerCmd(pp)
			}
			if cx, ok := cc.(interface{ SetRoot(root *RootCommand) }); ok {
				cx.SetRoot(root)
			}

			// } else {
			// 	// others: flags (if ff is not nil)
			// 	logz.VerboseContext(ctx, "    .|. ensureTreeR, (+owner,+root), Flags:", "flg", ff)
			// 	ff.SetOwnerCmd(cx)
			// 	ff.SetRoot(root)
		}
	})
}

// EnsureXref builds the internal indexes and maps.
//
// Called by worker.Worker in preparing time (preProcess).
//
// ForeachSubCommands, ForeachFlags, ForeachGroupedSubCommands, and
// ForeachGroupedFlags needs EnsureXref called.
func (c *CmdS) EnsureXref(ctx context.Context, cb ...func(cc Cmd, index, level int)) {
	c.Walk(ctx, func(cc Cmd, index, level int) {
		if cx, ok := cc.(*CmdS); ok {
			cx.ensureXrefCommands(ctx)
			cx.ensureXrefFlags(ctx)
			cx.ensureXrefGroup(ctx)
		}
		for _, fn := range cb {
			fn(cc, index, level)
		}
	})
}

func (c *CmdS) ensureXrefCommands(context.Context) { //nolint:revive
	if c.longCommands == nil {
		c.longCommands = make(map[string]*CmdS)
		for _, cc := range c.commands {
			for _, ss := range cc.GetLongTitleNamesArray() {
				c.longCommands[ss] = cc
			}
		}
	}
	if c.shortCommands == nil {
		c.shortCommands = make(map[string]*CmdS)
		for _, cc := range c.commands {
			for _, ss := range cc.GetShortTitleNamesArray() {
				c.shortCommands[ss] = cc
			}
		}
	}

	// if c.allCommands == nil {
	// 	c.allCommands = make(map[string]map[string]*CmdS)
	// 	for _, cc := range c.commands {
	// 		if cc.Short != "" {
	// 			c.shortCommands[cc.Short] = cc
	// 		}
	// 	}
	// }
}

func (c *CmdS) ensureXrefFlags(ctx context.Context) { //nolint:revive
	if c.longFlags == nil {
		c.longFlags = make(map[string]*Flag)
		for _, ff := range c.flags {
			c.ensureToggleGroups(ff)
			ff.ensureXref()
			if ff.headLike {
				if ff.owner.HeadLikeFlag() != nil && ff.owner.HeadLikeFlag() != ff {
					logz.WarnContext(ctx, "too much head-like flags", "last-head-like-flag", ff.owner.HeadLikeFlag(), "this-one", ff)
				}
				ff.owner.SetHeadLikeFlag(ff)
			}
			for _, ss := range ff.GetLongTitleNamesArray() {
				c.longFlags[ss] = ff
			}
		}
	}
	if c.shortFlags == nil {
		c.shortFlags = make(map[string]*Flag)
		for _, ff := range c.flags {
			c.ensureToggleGroups(ff)
			ff.ensureXref()
			if ff.headLike {
				if ff.owner.HeadLikeFlag() != nil && ff.owner.HeadLikeFlag() != ff {
					logz.WarnContext(ctx, "too much head-like flags", "last-head-like-flag", ff.owner.HeadLikeFlag(), "this-one", ff)
				}
				ff.owner.SetHeadLikeFlag(ff)
			}
			for _, ss := range ff.GetShortTitleNamesArray() {
				c.shortFlags[ss] = ff
			}
		}
	}
}

func (c *CmdS) ensureToggleGroups(ff *Flag) {
	if tg := ff.ToggleGroup(); tg != "" {
		if c.toggles == nil {
			c.toggles = make(map[string]*ToggleGroupMatch)
		}
		if c.toggles[tg] == nil {
			c.toggles[tg] = &ToggleGroupMatch{Flags: make(map[string]*Flag)}
		}
		c.toggles[tg].Flags[ff.Title()] = ff
		if ff.group == "" {
			ff.group = tg
		}
	}
}

func (c *CmdS) ensureXrefGroup(ctx context.Context) { //nolint:revive
	if c.allCommands == nil {
		c.allCommands = make(map[string]*CmdSlice)
		for _, cc := range c.commands {
			cc.ensureXrefCommands(ctx)
			cc.ensureXrefFlags(ctx)
			group := cc.SafeGroup()
			if m, ok := c.allCommands[group]; ok {
				m.A = append(m.A, cc)
			} else {
				c.allCommands[group] = &CmdSlice{A: []*CmdS{cc}}
			}
		}
	}
	if c.allFlags == nil {
		c.allFlags = make(map[string]*FlgSlice)
		for _, cc := range c.flags {
			group := cc.SafeGroup()
			if m, ok := c.allFlags[group]; ok {
				m.A = append(m.A, cc)
			} else {
				c.allFlags[group] = &FlgSlice{A: []*Flag{cc}}
			}
		}
	}
}

func (c *CmdS) invokeExternalEditor(ctx context.Context, vp *FlagValuePkg, ff *Flag) *Flag {
	if vp.Remains != "" {
		arg := c.normalizeStringValue(vp.Remains)
		vp.ValueOK, vp.Value, vp.Remains = true, arg, ""
		ff.defaultValue = arg
		return ff
	}
	if vp.AteArgs < len(vp.Args) {
		arg := c.normalizeStringValue(vp.Args[vp.AteArgs])
		if !strings.HasPrefix(arg, "-") && !strings.HasPrefix(arg, "~") {
			vp.ValueOK, vp.Value, vp.AteArgs = true, arg, vp.AteArgs+1
			ff.defaultValue = arg
			return ff
		}
	}

	logz.DebugContext(ctx, "external editor", "ex-editor", ff.externalEditor)
	if cmd := os.Getenv(ff.externalEditor); cmd != "" {
		file := tool.TempFileName("message*.tmp", "message001.tmp", c.App().Name())
		cmdS := tool.SplitCommandString(cmd)
		cmdS = append(cmdS, file)
		defer func(dst string) {
			if err := dir.DeleteFile(dst); err != nil {
				logz.ErrorContext(ctx, "cannot delete temporary file for flag", "flag", ff)
			}
		}(file)

		logz.DebugContext(ctx, "invoke external editor", "ex-editor", ff.externalEditor, "cmd", cmdS)
		if is.DebuggerAttached() {
			vp.ValueOK, vp.Value = true, "<<stdoutTextForDebugging>>"
			logz.WarnContext(ctx, "use debug text", "flag", ff, "text", vp.Value)
			return ff
		}

		if err := exec.CallSliceQuiet([]string{"which", cmdS[0]}, func(retCode int, stdoutText string) {
			if retCode == 0 {
				cmdS[0] = strings.TrimSpace(strings.TrimSuffix(stdoutText, "\n"))
				logz.DebugContext(ctx, "got external editor real-path", "cmd", cmdS)
			}
		}); err != nil {
			logz.ErrorContext(ctx, "cannot invoke which CmdS", "flag", ff, "cmd", cmdS)
			return nil
		}

		var content []byte
		var err error
		content, err = tool.LaunchEditorWithGetter(cmdS[0], func() string { return cmdS[1] }, false)
		if err != nil {
			logz.ErrorContext(ctx, "Error on launching cmd", "err", err, "cmd", cmdS)
			return nil
		}

		// content, err = tool.LaunchEditorWith(cmdS[0], cmdS[1])
		// if err != nil {
		// 	logz.ErrorContext(ctx, "Error on launching cmd", "err", err, "cmd", cmdS)
		// 	return nil
		// }
		//
		// content, err = tool.LaunchEditor(cmdS[0])
		// if err != nil {
		// 	logz.ErrorContext(ctx, "Error on launching cmd", "err", err, "cmd", cmdS)
		// 	return nil
		// }

		// f, err = os.Open(file)
		// if err != nil {
		// 	logz.ErrorContext(ctx, "cannot open temporary file for reading content", "file", file, "flag", ff, "cmd", cmdS)
		// 	return nil
		// }
		// defer f.Close()
		// vp.ValueOK, vp.Value = true, dir.MustReadAll(f)

		vp.ValueOK, vp.Value = true, string(content)
		ff.defaultValue = string(content)
		// logz.DebugContext(ctx, "invoked external editor", "ex-editor", ff.externalEditor, "text", string(content))
		return ff
	}
	logz.WarnContext(ctx, "Unknown External Editor for flag.", "ex-editor", ff.externalEditor, "flag", ff)
	return nil
}

// EqualTo compares with another one based on its titles
func (c *CmdS) EqualTo(rh *CmdS) (ok bool) {
	if c == nil {
		return rh == nil
	}
	if rh == nil {
		return false
	}
	return c.GetTitleName() == rh.GetTitleName()
}

func (c *CmdS) GetGroupedCommands(group string) (commands []*CmdS) {
	ctx := context.Background()
	c.ensureXrefGroup(ctx)
	commands = c.allCommands[group].A
	return
}

func (c *CmdS) GetGroupedFlags(group string) (flags []*Flag) {
	ctx := context.Background()
	c.ensureXrefGroup(ctx)
	flags = c.allFlags[group].A
	return
}

func (c *CmdS) CountOfCommands() int {
	vc := states.Env().CountOfVerbose()
	cnt := 0
	for _, cc := range c.commands {
		if cc.vendorHidden {
			if vc > 2 {
				cnt++
			}
		} else if cc.hidden {
			if vc > 0 {
				cnt++
			}
		} else {
			cnt++
		}
	}
	return cnt
}

func (c *CmdS) CountOfFlags() int {
	vc := states.Env().CountOfVerbose()
	cnt := 0
	for _, cc := range c.flags {
		// if (vc > 0 && cc.Hidden()) || (vc > 2 && cc.VendorHidden()) || (!cc.hidden && !cc.vendorHidden) {
		// 	cnt++
		// }
		if cc.vendorHidden {
			if vc > 2 {
				cnt++
			}
		} else if cc.hidden {
			if vc > 0 {
				cnt++
			}
		} else {
			cnt++
		}
	}
	return cnt
}
