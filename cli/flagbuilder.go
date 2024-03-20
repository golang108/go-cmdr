package cli

type FlagBuilder interface {
	// Build connects the built command into the building command system.
	Build()

	// Titles should be specified with this form:
	//
	//     longTitle, shortTitle, aliases...
	//
	// The Long-Title is must-required, and the others are optional.
	//
	// For Flag, Long-Title and Aliases are posix long parameters with the
	// leading double hyphen string '--'. And Short-Title has single
	// hyphen '-' as leading.
	//
	// For example, A flag with longTitle "debug" means that an end-user
	// should type "--debug" for it.
	//
	// For the multi-level command and subcommands, long, short and
	// aliases will be used as is.
	Titles(longTitle string, titles ...string) FlagBuilder
	// Default is a synonym to DefaultValue
	Default(defaultValue any) FlagBuilder

	// ExtraShorts sets more short titles
	ExtraShorts(shorts ...string) FlagBuilder

	// Description specifies the one-line description and a multi-line
	// description (optional)
	Description(description string, longDescription ...string) FlagBuilder
	// Examples can be a multi-line string.
	Examples(examples string) FlagBuilder
	// Group specify a group name,
	// A special prefix could sort it, has a form like `[0-9a-zA-Z]+\.`.
	// The prefix will be removed from help screen.
	//
	// Some examples are:
	//    "A001.Host Params"
	//    "A002.User Params"
	//
	// If ToggleGroup specified, Group field can be omitted because we will copy
	// from there.
	Group(group string) FlagBuilder
	// Deprecated is a version string just like '0.5.9' or 'v0.5.9', that
	// means this command/flag was/will be deprecated since `v0.5.9`.
	Deprecated(deprecated string) FlagBuilder
	// Hidden command/flag won't be shown in help-screen and others output.
	//
	// The Hidden command/flag may be printed normally if very verbose mode
	// specified (typically '-vv' detected).
	//
	// The VendorHidden commands/flags will be hidden at any time even if
	// in vert verbose mode.
	Hidden(hidden bool, vendorHidden ...bool) FlagBuilder

	ToggleGroup(group string) FlagBuilder
	PlaceHolder(placeHolder string) FlagBuilder

	// DefaultValue specifies a binding value to the flag with explicit
	// datatype.
	DefaultValue(val any) FlagBuilder
	// EnvVars binds the environment variable onto the flag
	EnvVars(vars ...string) FlagBuilder
	// AppendEnvVars binds the environment variable onto the flag
	AppendEnvVars(vars ...string) FlagBuilder
	// ExternalEditor is an env-var name to identify an external program
	// which will be used to collect user-input as a string value.
	//
	// The input string value will be bound to this value finally.
	ExternalEditor(externalEditor string) FlagBuilder
	// ValidArgs provides the selectable choice from a set of values.
	//
	// As end-user inputs not in the preset values, an error will be threw up.
	ValidArgs(validArgs ...string) FlagBuilder
	// AppendValidArgs provides the selectable choice from a set of values.
	//
	// As end-user inputs not in the preset values, an error will be threw up.
	AppendValidArgs(validArgs ...string) FlagBuilder
	// Range _
	// not yet
	Range(min, max int) FlagBuilder
	// HeadLike identifies this flag is head-like.
	HeadLike(headLike bool, bounds ...int) FlagBuilder
	// Required identifies this flag is must-required.
	//
	// As end-user input missed, an error will be threw up.
	Required(required bool) FlagBuilder

	// CompJustOnce is used for zsh completion.
	CompJustOnce(justOnce bool) FlagBuilder
	// CompActionStr is for zsh completion, see action of an optspec in _argument
	CompActionStr(action string) FlagBuilder
	// CompMutualExclusives is used for zsh completion.
	//
	// For the ToggleGroup group, mutualExclusives is implicit.
	CompMutualExclusives(ex ...string) FlagBuilder
	// CompPrerequisites flags for this one.
	//
	// In zsh completion, any of prerequisites flags must be present
	// so that user can complete this one.
	//
	// The prerequisites were not present and cmdr would report error
	// and stop parsing flow.
	CompPrerequisites(flags ...string) FlagBuilder
	// CompCircuitBreak is used for zsh completion.
	//
	// A flag can break cmdr parsing flow with return
	// ErrShouldBeStopException in its Action handler.
	// But you' better told zsh system with set circuitBreak
	// to true. At this case, cmdr will generate a suitable
	// completion script.
	CompCircuitBreak(cb bool) FlagBuilder

	// DoubleTildeOnly can be used for zsh completion.
	//
	// A DoubleTildeOnly Flag accepts '~~opt' only, so '--opt' is
	// invalid form and couldn't be used for other Flag
	// anymore.
	DoubleTildeOnly(b bool) FlagBuilder

	// OnParseValue allows user-defined value parsing, converting and validating.
	OnParseValue(handler OnParseValueHandler) FlagBuilder
	// OnMatched handler will be called when this flag matched.
	//
	// OnMatched handler is a cancellable notifier (a validator)
	// before a formal on-changed notification,
	//
	// OnMatched will be called after a flag matched and its value
	// extracted but not saved.
	//
	// = OnValidating
	//
	// You can capture it and validate the user input for this flag.
	//
	// If you're looking for a best hook point where the old value is
	// changing to new value, using OnChanging handler.
	//
	// The calling order in parsing command-line:
	//
	//     OnParseValue (cancel ->)
	//     OnMatched    (cancel ->)
	//     OnChanging   (cancel ->)
	//     OnChanged
	//
	// The calling order in parsing other sources (config file, ...):
	//
	//     OnParseValue (cancel ->)
	//     OnMatched    (cancel ->)
	//     OnChanging   (cancel ->)
	//     OnChanged
	//
	// The calling order if store.Set(dottedPath, value) calling:
	//
	//     OnParseValue (cancel ->)
	//     OnSet
	//
	OnMatched(handler OnMatchedHandler) FlagBuilder
	OnChanging(handler OnChangingHandler) FlagBuilder
	// OnChanged handler will be called when this flag is being
	// modified generally (programmatically, cmdline parsing, cfg file, ...)
	OnChanged(handler OnChangedHandler) FlagBuilder
	// OnSet handler will be called when this flag is being modified
	// programmatically.
	OnSet(handler OnSetHandler) FlagBuilder
}
