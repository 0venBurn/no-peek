package main

type appMode string

const (
	modePuzzle appMode = "puzzle"
	modeDeep   appMode = "deep"
)

type menuFieldID string

const (
	fieldMode       menuFieldID = "mode"
	fieldSession    menuFieldID = "session"
	fieldFocus      menuFieldID = "focus"
	fieldRescue     menuFieldID = "rescue"
	fieldDeepFocus  menuFieldID = "deepFocus"
	fieldShortBreak menuFieldID = "shortBreak"
	fieldLongBreak  menuFieldID = "longBreak"
	fieldBreaks     menuFieldID = "breaks"
	fieldCycles     menuFieldID = "cycles"
)

type menuFieldKind int

const (
	menuFieldMode menuFieldKind = iota
	menuFieldText
	menuFieldMinutes
	menuFieldToggle
	menuFieldNumber
)

type menuFieldDef struct {
	id    menuFieldID
	label string
	kind  menuFieldKind
}

type modeDefinition struct {
	mode           appMode
	defaultProblem string
	menuFields     []menuFieldDef
}

var puzzleModeDefinition = modeDefinition{
	mode:           modePuzzle,
	defaultProblem: "Untitled problem",
	menuFields: []menuFieldDef{
		{id: fieldMode, label: "Mode", kind: menuFieldMode},
		{id: fieldSession, label: "Session", kind: menuFieldText},
		{id: fieldFocus, label: "Focus", kind: menuFieldMinutes},
		{id: fieldRescue, label: "Rescue", kind: menuFieldMinutes},
	},
}

var deepModeDefinition = modeDefinition{
	mode:           modeDeep,
	defaultProblem: "Deep work",
	menuFields: []menuFieldDef{
		{id: fieldMode, label: "Mode", kind: menuFieldMode},
		{id: fieldSession, label: "Session", kind: menuFieldText},
		{id: fieldDeepFocus, label: "Focus", kind: menuFieldMinutes},
		{id: fieldShortBreak, label: "Short break", kind: menuFieldMinutes},
		{id: fieldLongBreak, label: "Long break", kind: menuFieldMinutes},
		{id: fieldBreaks, label: "Breaks", kind: menuFieldToggle},
		{id: fieldCycles, label: "Cycles", kind: menuFieldNumber},
	},
}

func modeDef(mode appMode) modeDefinition {
	if mode == modeDeep {
		return deepModeDefinition
	}
	return puzzleModeDefinition
}

func validMode(mode appMode) bool {
	return mode == modePuzzle || mode == modeDeep
}

func menuFieldsFor(mode appMode) []menuFieldDef {
	return modeDef(mode).menuFields
}

func defaultProblem(mode appMode) string {
	return modeDef(mode).defaultProblem
}
