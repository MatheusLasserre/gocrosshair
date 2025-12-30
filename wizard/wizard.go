// Package wizard provides an interactive terminal UI for configuring gocrosshair.
package wizard

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"gocrosshair/config"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			MarginBottom(1)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("229")).
			Bold(true)

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("78")).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginTop(1)
)

type step int

const (
	stepShape step = iota
	stepColor
	stepSize
	stepThickness
	stepGap
	stepOutline
	stepOutlineColor
	stepMonitor
	stepOffsetX
	stepOffsetY
	stepConfirm
	stepStartPrompt
	stepDone
)

type Monitor struct {
	Index   int
	Name    string
	Width   uint16
	Height  uint16
	Primary bool
}

// Model is the Bubble Tea model for the setup wizard
type Model struct {
	step           step
	config         *config.Config
	monitors       []Monitor
	configPath     string
	cursor         int
	textInput      textinput.Model
	err            error
	quitting       bool
	saved          bool
	startCrosshair bool
	width, height  int
}

var shapeOptions = []string{"cross", "dot", "circle", "cross-dot"}

var colorPresets = []struct {
	name  string
	value string
}{
	{"Green", "#00FF00"},
	{"Red", "#FF0000"},
	{"White", "#FFFFFF"},
	{"Cyan", "#00FFFF"},
	{"Yellow", "#FFFF00"},
	{"Pink", "#FF00FF"},
	{"Orange", "#FFA500"},
	{"Custom...", "custom"},
}

// NewModel creates a new wizard model
func NewModel(monitors []Monitor, configPath string) Model {
	ti := textinput.New()
	ti.Focus()
	ti.CharLimit = 20
	ti.Width = 20

	return Model{
		step:       stepShape,
		config:     config.Default(),
		monitors:   monitors,
		configPath: configPath,
		cursor:     0,
		textInput:  ti,
	}
}

// Init implements tea.Model
func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

// Update implements tea.Model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			m.cursor = min(m.cursor+1, m.maxCursor())

		case "enter":
			return m.handleEnter()

		case "esc":
			if m.step > stepShape {
				m.step--
				// Skip thickness and gap when going back for dot/circle shapes
				if !m.shapeNeedsThicknessAndGap() {
					if m.step == stepGap || m.step == stepThickness {
						m.step = stepSize
						m.textInput.SetValue(strconv.Itoa(m.config.Crosshair.Size))
						m.textInput.Focus()
						m.cursor = 0
						m.err = nil
						return m, textinput.Blink
					}
				}
				m.cursor = 0
				m.err = nil
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	// Handle text input for custom values
	if m.isTextInputStep() {
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

// View implements tea.Model
func (m Model) View() string {
	if m.quitting {
		if m.saved {
			if m.startCrosshair {
				return successStyle.Render("✓ Configuration saved! Starting crosshair...") + "\n"
			}
			return successStyle.Render("✓ Configuration saved to "+m.configPath) + "\n"
		}
		return dimStyle.Render("Configuration cancelled.") + "\n"
	}

	var b strings.Builder

	b.WriteString(titleStyle.Render("🎯 gocrosshair Setup Wizard") + "\n\n")

	switch m.step {
	case stepShape:
		b.WriteString(m.renderSelect("Select crosshair shape:", shapeOptions))

	case stepColor:
		b.WriteString(m.renderColorSelect())

	case stepSize:
		b.WriteString(m.renderNumberInput("Crosshair size (pixels from center):", "10", 1, 100))

	case stepThickness:
		b.WriteString(m.renderNumberInput("Line thickness (pixels):", "2", 1, 20))

	case stepGap:
		b.WriteString(m.renderNumberInput("Center gap (0 for solid, or pixels):", "0", 0, 50))

	case stepOutline:
		b.WriteString(m.renderSelect("Add outline?", []string{"No", "Yes"}))

	case stepOutlineColor:
		b.WriteString(m.renderColorSelect())

	case stepMonitor:
		b.WriteString(m.renderMonitorSelect())

	case stepOffsetX:
		b.WriteString(m.renderNumberInput("Horizontal offset from center (pixels, negative=left):", "0", -500, 500))

	case stepOffsetY:
		b.WriteString(m.renderNumberInput("Vertical offset from center (pixels, negative=up):", "0", -500, 500))

	case stepConfirm:
		b.WriteString(m.renderConfirm())

	case stepStartPrompt:
		b.WriteString(m.renderStartPrompt())

	case stepDone:
		b.WriteString(successStyle.Render("✓ Configuration saved!") + "\n")
		return b.String()
	}

	if m.err != nil {
		b.WriteString("\n" + errorStyle.Render("Error: "+m.err.Error()))
	}

	b.WriteString(helpStyle.Render("\n↑/↓ navigate • enter select • esc back • q quit"))

	return b.String()
}

func (m Model) renderSelect(title string, options []string) string {
	var b strings.Builder
	b.WriteString(normalStyle.Render(title) + "\n\n")

	for i, opt := range options {
		cursor := "  "
		style := normalStyle
		if i == m.cursor {
			cursor = "▸ "
			style = selectedStyle
		}
		b.WriteString(cursor + style.Render(opt) + "\n")
	}

	return b.String()
}

func (m Model) renderColorSelect() string {
	var b strings.Builder

	if m.step == stepColor {
		b.WriteString(normalStyle.Render("Select crosshair color:") + "\n\n")
	} else {
		b.WriteString(normalStyle.Render("Select outline color:") + "\n\n")
	}

	// Check if we're in custom input mode
	if m.cursor == len(colorPresets)-1 && m.textInput.Focused() {
		for i, preset := range colorPresets[:len(colorPresets)-1] {
			b.WriteString("  " + normalStyle.Render(preset.name) + dimStyle.Render(" "+preset.value) + "\n")
			_ = i
		}
		b.WriteString("▸ " + selectedStyle.Render("Custom: ") + m.textInput.View() + "\n")
		b.WriteString(dimStyle.Render("  (Enter hex color like #FF0000)") + "\n")
		return b.String()
	}

	for i, preset := range colorPresets {
		cursor := "  "
		style := normalStyle
		if i == m.cursor {
			cursor = "▸ "
			style = selectedStyle
		}
		colorPreview := ""
		if preset.value != "custom" {
			colorPreview = dimStyle.Render(" " + preset.value)
		}
		b.WriteString(cursor + style.Render(preset.name) + colorPreview + "\n")
	}

	return b.String()
}

func (m Model) renderMonitorSelect() string {
	var b strings.Builder
	b.WriteString(normalStyle.Render("Select monitor:") + "\n\n")

	// Add "Primary (auto)" option
	cursor := "  "
	style := normalStyle
	if m.cursor == 0 {
		cursor = "▸ "
		style = selectedStyle
	}
	b.WriteString(cursor + style.Render("Primary (auto-detect)") + "\n")

	for i, mon := range m.monitors {
		cursor := "  "
		style := normalStyle
		if i+1 == m.cursor {
			cursor = "▸ "
			style = selectedStyle
		}
		primary := ""
		if mon.Primary {
			primary = dimStyle.Render(" ← primary")
		}
		desc := fmt.Sprintf("[%d] %s: %dx%d", mon.Index, mon.Name, mon.Width, mon.Height)
		b.WriteString(cursor + style.Render(desc) + primary + "\n")
	}

	return b.String()
}

func (m Model) renderNumberInput(title, defaultVal string, minVal, maxVal int) string {
	var b strings.Builder
	b.WriteString(normalStyle.Render(title) + "\n\n")
	b.WriteString("▸ " + m.textInput.View() + "\n")
	b.WriteString(dimStyle.Render(fmt.Sprintf("  (Range: %d to %d, default: %s)", minVal, maxVal, defaultVal)) + "\n")
	return b.String()
}

func (m Model) renderConfirm() string {
	var b strings.Builder
	b.WriteString(normalStyle.Render("Review your configuration:") + "\n\n")

	cfg := m.config
	b.WriteString(dimStyle.Render("  Shape:     ") + normalStyle.Render(cfg.Crosshair.Shape) + "\n")
	b.WriteString(dimStyle.Render("  Color:     ") + normalStyle.Render(cfg.Crosshair.Color) + "\n")
	b.WriteString(dimStyle.Render("  Size:      ") + normalStyle.Render(fmt.Sprintf("%d px", cfg.Crosshair.Size)) + "\n")
	// Only show thickness and gap for cross/cross-dot shapes
	if m.shapeNeedsThicknessAndGap() {
		b.WriteString(dimStyle.Render("  Thickness: ") + normalStyle.Render(fmt.Sprintf("%d px", cfg.Crosshair.Thickness)) + "\n")
		b.WriteString(dimStyle.Render("  Gap:       ") + normalStyle.Render(fmt.Sprintf("%d px", cfg.Crosshair.Gap)) + "\n")
	}
	if cfg.Crosshair.OutlineThickness > 0 {
		b.WriteString(dimStyle.Render("  Outline:   ") + normalStyle.Render(fmt.Sprintf("%d px (%s)", cfg.Crosshair.OutlineThickness, cfg.Crosshair.OutlineColor)) + "\n")
	}
	if cfg.Position.Monitor == -1 {
		b.WriteString(dimStyle.Render("  Monitor:   ") + normalStyle.Render("Primary (auto)") + "\n")
	} else {
		b.WriteString(dimStyle.Render("  Monitor:   ") + normalStyle.Render(fmt.Sprintf("%d", cfg.Position.Monitor)) + "\n")
	}
	if cfg.Position.OffsetX != 0 || cfg.Position.OffsetY != 0 {
		b.WriteString(dimStyle.Render("  Offset:    ") + normalStyle.Render(fmt.Sprintf("(%d, %d)", cfg.Position.OffsetX, cfg.Position.OffsetY)) + "\n")
	}

	b.WriteString("\n")
	options := []string{"Save configuration", "Start over", "Cancel"}
	for i, opt := range options {
		cursor := "  "
		style := normalStyle
		if i == m.cursor {
			cursor = "▸ "
			style = selectedStyle
		}
		b.WriteString(cursor + style.Render(opt) + "\n")
	}

	return b.String()
}

func (m Model) renderStartPrompt() string {
	var b strings.Builder
	b.WriteString(successStyle.Render("✓ Configuration saved!") + "\n\n")
	b.WriteString(normalStyle.Render("Start the crosshair now?") + "\n\n")

	options := []string{"Yes, start now", "No, exit"}
	for i, opt := range options {
		cursor := "  "
		style := normalStyle
		if i == m.cursor {
			cursor = "▸ "
			style = selectedStyle
		}
		b.WriteString(cursor + style.Render(opt) + "\n")
	}

	return b.String()
}

func (m Model) maxCursor() int {
	switch m.step {
	case stepShape:
		return len(shapeOptions) - 1
	case stepColor, stepOutlineColor:
		return len(colorPresets) - 1
	case stepOutline:
		return 1
	case stepMonitor:
		return len(m.monitors)
	case stepConfirm:
		return 2
	case stepStartPrompt:
		return 1
	default:
		return 0
	}
}

func (m Model) isTextInputStep() bool {
	switch m.step {
	case stepSize, stepThickness, stepGap, stepOffsetX, stepOffsetY:
		return true
	case stepColor, stepOutlineColor:
		return m.cursor == len(colorPresets)-1
	default:
		return false
	}
}

// shapeNeedsThicknessAndGap returns true if the current shape needs thickness and gap settings
func (m Model) shapeNeedsThicknessAndGap() bool {
	shape := m.config.Crosshair.Shape
	return shape == "cross" || shape == "cross-dot"
}

func (m Model) handleEnter() (tea.Model, tea.Cmd) {
	switch m.step {
	case stepShape:
		m.config.Crosshair.Shape = shapeOptions[m.cursor]
		m.step = stepColor
		m.cursor = 0

	case stepColor:
		if m.cursor == len(colorPresets)-1 {
			if !m.textInput.Focused() {
				m.textInput.SetValue("")
				m.textInput.Focus()
				return m, textinput.Blink
			}
			color := m.textInput.Value()
			if !strings.HasPrefix(color, "#") {
				color = "#" + color
			}
			if _, err := config.ParseColor(color); err != nil {
				m.err = fmt.Errorf("invalid color format")
				return m, nil
			}
			m.config.Crosshair.Color = color
			m.textInput.Blur()
		} else {
			m.config.Crosshair.Color = colorPresets[m.cursor].value
		}
		m.step = stepSize
		m.cursor = 0
		m.textInput.SetValue("10")
		m.textInput.Focus()
		m.err = nil

	case stepSize:
		val, err := strconv.Atoi(m.textInput.Value())
		if err != nil || val < 1 || val > 100 {
			m.err = fmt.Errorf("enter a number between 1 and 100")
			return m, nil
		}
		m.config.Crosshair.Size = val
		// Skip thickness and gap for dot/circle shapes
		if m.shapeNeedsThicknessAndGap() {
			m.step = stepThickness
			m.textInput.SetValue("2")
		} else {
			m.step = stepOutline
			m.cursor = 0
			m.textInput.Blur()
		}
		m.err = nil

	case stepThickness:
		val, err := strconv.Atoi(m.textInput.Value())
		if err != nil || val < 1 || val > 20 {
			m.err = fmt.Errorf("enter a number between 1 and 20")
			return m, nil
		}
		m.config.Crosshair.Thickness = val
		m.step = stepGap
		m.textInput.SetValue("0")
		m.err = nil

	case stepGap:
		val, err := strconv.Atoi(m.textInput.Value())
		if err != nil || val < 0 || val > 50 {
			m.err = fmt.Errorf("enter a number between 0 and 50")
			return m, nil
		}
		m.config.Crosshair.Gap = val
		m.step = stepOutline
		m.cursor = 0
		m.textInput.Blur()
		m.err = nil

	case stepOutline:
		if m.cursor == 0 {
			m.config.Crosshair.OutlineThickness = 0
			m.step = stepMonitor
		} else {
			m.config.Crosshair.OutlineThickness = 1
			m.step = stepOutlineColor
		}
		m.cursor = 0
		m.err = nil

	case stepOutlineColor:
		if m.cursor == len(colorPresets)-1 {
			if !m.textInput.Focused() {
				m.textInput.SetValue("")
				m.textInput.Focus()
				return m, textinput.Blink
			}
			color := m.textInput.Value()
			if !strings.HasPrefix(color, "#") {
				color = "#" + color
			}
			if _, err := config.ParseColor(color); err != nil {
				m.err = fmt.Errorf("invalid color format")
				return m, nil
			}
			m.config.Crosshair.OutlineColor = color
			m.textInput.Blur()
		} else {
			m.config.Crosshair.OutlineColor = colorPresets[m.cursor].value
		}
		m.step = stepMonitor
		m.cursor = 0
		m.err = nil

	case stepMonitor:
		if m.cursor == 0 {
			m.config.Position.Monitor = -1
		} else {
			m.config.Position.Monitor = m.cursor - 1
		}
		m.step = stepOffsetX
		m.textInput.SetValue("0")
		m.textInput.Focus()
		m.err = nil

	case stepOffsetX:
		val, err := strconv.Atoi(m.textInput.Value())
		if err != nil || val < -500 || val > 500 {
			m.err = fmt.Errorf("enter a number between -500 and 500")
			return m, nil
		}
		m.config.Position.OffsetX = val
		m.step = stepOffsetY
		m.textInput.SetValue("0")
		m.err = nil

	case stepOffsetY:
		val, err := strconv.Atoi(m.textInput.Value())
		if err != nil || val < -500 || val > 500 {
			m.err = fmt.Errorf("enter a number between -500 and 500")
			return m, nil
		}
		m.config.Position.OffsetY = val
		m.step = stepConfirm
		m.cursor = 0
		m.textInput.Blur()
		m.err = nil

	case stepConfirm:
		switch m.cursor {
		case 0:
			if err := config.Save(m.configPath, m.config); err != nil {
				m.err = err
				return m, nil
			}
			m.saved = true
			m.step = stepStartPrompt
			m.cursor = 0
		case 1:
			m.config = config.Default()
			m.step = stepShape
			m.cursor = 0
		case 2:
			m.quitting = true
			return m, tea.Quit
		}

	case stepStartPrompt:
		if m.cursor == 0 {
			m.startCrosshair = true
		}
		m.quitting = true
		return m, tea.Quit
	}

	return m, nil
}

// GetConfig returns the configured settings
func (m Model) GetConfig() *config.Config {
	return m.config
}

// WasSaved returns true if the configuration was saved
func (m Model) WasSaved() bool {
	return m.saved
}

// WantsToStart returns true if the user wants to start the crosshair
func (m Model) WantsToStart() bool {
	return m.startCrosshair
}
