package core

import (
	"bytes"
	"time"

	"catalyst/internal/app/styles"

	"github.com/charmbracelet/bubbles/v2/viewport"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/charmbracelet/log/v2"
)

type LogsViewModel struct {
	viewport  viewport.Model
	logger    *log.Logger
	logOutput *bytes.Buffer
	width     int
	height    int
	theme     *styles.Theme
}

func NewLogsView(
	width, availableHeight int,
	theme *styles.Theme,
) *LogsViewModel {
	logOutput := new(bytes.Buffer)
	logger := log.NewWithOptions(logOutput, log.Options{
		ReportCaller:    false,
		ReportTimestamp: true,
		TimeFormat:      time.Kitchen,
	})
	logger.SetLevel(log.DebugLevel)
	logger.SetColorProfile(colorprofile.TrueColor)
	style := log.DefaultStyles()
	style.Key = style.Key.Foreground(theme.Accent)
	style.Value = style.Value.Foreground(theme.FgHalfMuted)
	logger.SetStyles(style)

	vp := viewport.New()
	vp.SetHeight(availableHeight)
	vp.SetWidth(width)

	logger.Debug("Initializing log view...")

	vp.SetContent(logOutput.String())

	return &LogsViewModel{
		viewport:  vp,
		logger:    logger,
		logOutput: logOutput,
		width:     width,
		height:    availableHeight,
		theme:     theme,
	}
}

func (m *LogsViewModel) Init() tea.Cmd {
	return nil
}

func (m *LogsViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *LogsViewModel) View() string {
	m.viewport.Style = lipgloss.NewStyle().
		Padding(0, 1)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Left,
		lipgloss.Top,
		m.viewport.View(),
	)
}

func (m *LogsViewModel) Resize(width, availableHeight int) {
	m.width = width
	m.height = availableHeight
	m.viewport.SetWidth(width)
	m.viewport.SetHeight(availableHeight)
}

// AddLog adds a new line to the logs view with a specific log level and key-value pairs.
func (m *LogsViewModel) AddLog(level log.Level, msg string, keyvals ...any) {
	switch level {
	case log.DebugLevel:
		m.logger.Debug(msg, keyvals...)
	case log.InfoLevel:
		m.logger.Info(msg, keyvals...)
	case log.WarnLevel:
		m.logger.Warn(msg, keyvals...)
	case log.ErrorLevel:
		m.logger.Error(msg, keyvals...)
	case log.FatalLevel:
		m.logger.Fatal(msg, keyvals...)
	}
	m.viewport.SetContent(m.logOutput.String())
	m.viewport.GotoBottom()
}
