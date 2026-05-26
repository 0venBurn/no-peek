package main

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type menuState struct {
	cfg          config
	cursor       int
	editingField menuFieldID
	input        string
}

type menuKeyOutcome struct {
	menu  menuState
	start bool
	quit  bool
}

type menuRender struct {
	rows   []menuRowRender
	footer string
}

type menuRowRender struct {
	label    string
	value    string
	selected bool
	editing  bool
}

func newMenuState(cfg config) menuState {
	return menuState{cfg: normalizeConfig(cfg)}
}

func (m menuState) config() config {
	return normalizeConfig(m.cfg)
}

func (m menuState) handleKey(msg tea.KeyMsg) menuKeyOutcome {
	if m.editingField == "" {
		switch msg.String() {
		case "ctrl+c", "q":
			return menuKeyOutcome{menu: m, quit: true}
		case "s":
			return menuKeyOutcome{menu: m, start: true}
		case "enter":
			m.startEditingSelected()
		case "up", "k", "shift+tab":
			m.moveCursor(-1)
		case "down", "j", "tab":
			m.moveCursor(1)
		}
		return menuKeyOutcome{menu: m}
	}

	field := m.editingFieldDef()
	switch msg.String() {
	case "ctrl+c":
		return menuKeyOutcome{menu: m, quit: true}
	case "enter":
		m.commitEdit()
	case "esc":
		m.stopEditing()
	case "left":
		if field.kind == menuFieldMode || field.kind == menuFieldToggle || field.kind == menuFieldNumber {
			m.adjustSelected(-1)
		}
	case "h":
		switch field.kind {
		case menuFieldText:
			m.appendProblemInput("h")
		case menuFieldMode, menuFieldToggle, menuFieldNumber:
			m.adjustSelected(-1)
		}
	case "right":
		if field.kind == menuFieldMode || field.kind == menuFieldToggle || field.kind == menuFieldNumber {
			m.adjustSelected(1)
		}
	case "l":
		switch field.kind {
		case menuFieldText:
			m.appendProblemInput("l")
		case menuFieldMode, menuFieldToggle, menuFieldNumber:
			m.adjustSelected(1)
		}
	case " ":
		switch field.kind {
		case menuFieldMode:
			m.toggleMode()
		case menuFieldToggle:
			m.toggleBreaks()
		case menuFieldText:
			m.appendProblemInput(" ")
		}
	case "backspace":
		if field.kind == menuFieldText && len(m.cfg.problem) > 0 {
			m.cfg.problem = m.cfg.problem[:len(m.cfg.problem)-1]
		} else if (field.kind == menuFieldMinutes || field.kind == menuFieldNumber) && len(m.input) > 0 {
			m.input = m.input[:len(m.input)-1]
		}
	default:
		s := msg.String()
		if field.kind == menuFieldText && isPrintableRune(s) {
			m.appendProblemInput(s)
		} else if field.kind == menuFieldMinutes && isDigitRune(s) && len(m.input) < 3 {
			m.input += s
		} else if field.kind == menuFieldNumber && isDigitRune(s) && len(m.input) < 2 {
			m.input += s
		}
	}
	return menuKeyOutcome{menu: m}
}

func (m menuState) render() menuRender {
	fields := m.fields()
	rows := make([]menuRowRender, 0, len(fields))
	selected := m.selectedField().id
	for _, field := range fields {
		rows = append(rows, menuRowRender{
			label:    field.label,
			value:    m.fieldValue(field),
			selected: selected == field.id,
			editing:  m.editingField == field.id,
		})
	}
	return menuRender{rows: rows, footer: m.footerText()}
}

func (m menuState) footerText() string {
	if m.editingField == "" {
		return "j/k select · enter edit · s start · q quit"
	}
	switch m.editingFieldDef().kind {
	case menuFieldMinutes:
		return "type minutes, e.g. 25 · enter save · esc cancel"
	case menuFieldNumber:
		return "type cycles 1-99 · h/l adjust · enter save · esc cancel"
	default:
		return "type value or h/l toggle · enter done · esc cancel"
	}
}

func (m menuState) fields() []menuFieldDef {
	return menuFieldsFor(m.cfg)
}

func (m menuState) selectedField() menuFieldDef {
	fields := m.fields()
	if m.cursor < 0 || m.cursor >= len(fields) {
		return fields[0]
	}
	return fields[m.cursor]
}

func (m menuState) editingFieldDef() menuFieldDef {
	if field, ok := m.fieldByID(m.editingField); ok {
		return field
	}
	return m.selectedField()
}

func (m menuState) fieldByID(id menuFieldID) (menuFieldDef, bool) {
	for _, field := range m.fields() {
		if field.id == id {
			return field, true
		}
	}
	return menuFieldDef{}, false
}

func (m *menuState) moveCursor(delta int) {
	fields := m.fields()
	m.cursor = (m.cursor + delta + len(fields)) % len(fields)
	m.stopEditing()
}

func (m *menuState) startEditingSelected() {
	field := m.selectedField()
	m.editingField = field.id
	m.input = ""
	if field.kind == menuFieldText && m.cfg.problem == defaultProblem(m.cfg.mode) {
		m.cfg.problem = ""
	}
}

func (m *menuState) commitEdit() {
	field := m.editingFieldDef()
	if field.kind == menuFieldMinutes && m.input != "" {
		m.setMinuteField(field.id, parsePositiveInt(m.input))
	}
	if field.kind == menuFieldNumber && m.input != "" {
		m.cfg.deepCycles = clampDeepCycles(parsePositiveInt(m.input))
	}
	m.cfg = normalizeConfig(m.cfg)
	m.stopEditing()
}

func (m *menuState) stopEditing() {
	m.editingField = ""
	m.input = ""
}

func (m *menuState) adjustSelected(delta int) {
	field := m.selectedField()
	switch field.id {
	case fieldMode:
		m.toggleMode()
	case fieldFocus, fieldRescue, fieldDeepFocus, fieldShortBreak, fieldLongBreak:
		m.setMinuteField(field.id, m.minuteField(field.id)+delta)
	case fieldBreaks:
		m.toggleBreaks()
	case fieldCycles:
		m.cfg.deepCycles = clampDeepCycles(m.cfg.deepCycles + delta)
	}
}

func (m *menuState) toggleMode() {
	if m.cfg.mode == modePuzzle {
		m.cfg.mode = modeDeep
	} else {
		m.cfg.mode = modePuzzle
	}
	if m.cursor >= len(m.fields()) {
		m.cursor = len(m.fields()) - 1
	}
	if _, ok := m.fieldByID(m.editingField); !ok {
		m.stopEditing()
	}
}

func (m *menuState) toggleBreaks() {
	m.cfg.noBreaks = !m.cfg.noBreaks
	m.moveCursorToField(fieldBreaks)
}

func (m *menuState) moveCursorToField(id menuFieldID) {
	for i, field := range m.fields() {
		if field.id == id {
			m.cursor = i
			return
		}
	}
	if m.cursor >= len(m.fields()) {
		m.cursor = len(m.fields()) - 1
	}
}

func (m *menuState) appendProblemInput(s string) {
	if m.input == "" && m.cfg.problem == defaultProblem(m.cfg.mode) {
		m.cfg.problem = ""
	}
	m.cfg.problem += s
	m.input = "typed"
}

func (m menuState) minuteField(id menuFieldID) int {
	switch id {
	case fieldFocus:
		return wholeMinutes(m.cfg.focusDuration)
	case fieldRescue:
		return wholeMinutes(m.cfg.rescueDuration)
	case fieldDeepFocus:
		return wholeMinutes(m.cfg.deepFocusDuration)
	case fieldShortBreak:
		return wholeMinutes(m.cfg.shortBreakDuration)
	case fieldLongBreak:
		return wholeMinutes(m.cfg.longBreakDuration)
	default:
		return 1
	}
}

func (m *menuState) setMinuteField(id menuFieldID, minutes int) {
	if minutes < 1 {
		minutes = 1
	}
	d := time.Duration(minutes) * time.Minute
	switch id {
	case fieldFocus:
		m.cfg.focusDuration = d
	case fieldRescue:
		m.cfg.rescueDuration = d
	case fieldDeepFocus:
		m.cfg.deepFocusDuration = d
	case fieldShortBreak:
		m.cfg.shortBreakDuration = d
	case fieldLongBreak:
		m.cfg.longBreakDuration = d
	}
}

func (m menuState) fieldValue(field menuFieldDef) string {
	switch field.kind {
	case menuFieldMode:
		return string(m.cfg.mode)
	case menuFieldText:
		return m.cfg.problem
	case menuFieldMinutes:
		if m.editingField == field.id {
			if m.input == "" {
				return ""
			}
			return m.input + " min"
		}
		return formatMinutes(m.minuteField(field.id))
	case menuFieldToggle:
		if m.cfg.noBreaks {
			return "off"
		}
		return "on"
	case menuFieldNumber:
		if m.editingField == field.id && m.input != "" {
			return m.input
		}
		return formatInt(m.cfg.deepCycles)
	default:
		return ""
	}
}

func parsePositiveInt(s string) int {
	value := 0
	for _, r := range s {
		if r < '0' || r > '9' {
			continue
		}
		value = value*10 + int(r-'0')
	}
	if value < 1 {
		return 1
	}
	return value
}

func formatMinutes(minutes int) string {
	return fmt.Sprintf("%d min", minutes)
}

func formatInt(value int) string {
	return fmt.Sprintf("%d", value)
}

func isPrintableRune(s string) bool {
	return len(s) == 1 && s[0] >= 32 && s[0] <= 126
}

func isDigitRune(s string) bool {
	return len(s) == 1 && s[0] >= '0' && s[0] <= '9'
}
