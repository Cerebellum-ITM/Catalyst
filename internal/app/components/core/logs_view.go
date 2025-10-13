package core

import (
	"bytes"

	"catalyst/internal/app/styles"

	"github.com/charmbracelet/bubbles/v2/viewport"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/charmbracelet/log"
	"github.com/muesli/termenv"
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
	logger := log.New(logOutput)
	logger.SetColorProfile(termenv.TrueColor)
	logger.SetLevel(log.DebugLevel)

	vp := viewport.New()
	vp.SetHeight(availableHeight)
	vp.SetWidth(width)

	// Dummy messages
	logger.Info("Initializing log view...")
	logger.Info("Loading resources...")
	logger.Warn("This is a dummy warning message.")
	logger.Info("Ready.")

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
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.Accent).
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
