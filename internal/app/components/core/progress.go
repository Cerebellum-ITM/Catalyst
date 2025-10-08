package core

import (
	"fmt"
	"image/color"
	"math"
	"strings"
	"sync/atomic"
	"time"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/harmonica"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/lucasb-eyer/go-colorful"
)

var lastID int64

func nextID() int {
	return int(atomic.AddInt64(&lastID, 1))
}

const (
	fps              = 60
	defaultWidth     = 40
	defaultFrequency = 18.0
	defaultDamping   = 1.0
)

type ProgressOption func(*ProgressModel)

func WithDefaultGradient() ProgressOption {
	return WithGradient("#5A56E0", "#EE6FF8")
}

func WithGradient(colorA, colorB string) ProgressOption {
	return func(m *ProgressModel) {
		m.setRamp(colorA, colorB, false)
	}
}

func WithDefaultScaledGradient() ProgressOption {
	return WithScaledGradient("#5A56E0", "#EE6FF8")
}

func WithScaledGradient(colorA, colorB string) ProgressOption {
	return func(m *ProgressModel) {
		m.setRamp(colorA, colorB, true)
	}
}

func WithSolidFill(color color.Color) ProgressOption {
	return func(m *ProgressModel) {
		m.FullColor = color
		m.useRamp = false
	}
}

func WithFillCharacters(full rune, empty rune) ProgressOption {
	return func(m *ProgressModel) {
		m.Full = full
		m.Empty = empty
	}
}

func WithoutPercentage() ProgressOption {
	return func(m *ProgressModel) {
		m.ShowPercentage = false
	}
}

func WithWidth(w int) ProgressOption {
	return func(m *ProgressModel) {
		m.width = w
	}
}

func WithSpringOptions(frequency, damping float64) ProgressOption {
	return func(m *ProgressModel) {
		m.SetSpringOptions(frequency, damping)
		m.springCustomized = true
	}
}

type FrameMsg struct {
	id  int
	tag int
}

type ProgressModel struct {
	id               int
	tag              int
	width            int
	Full             rune
	FullColor        color.Color
	Empty            rune
	EmptyColor       color.Color
	ShowPercentage   bool
	PercentFormat    string
	PercentageStyle  lipgloss.Style
	spring           harmonica.Spring
	springCustomized bool
	percentShown     float64
	targetPercent    float64
	velocity         float64
	useRamp          bool
	rampColorA       colorful.Color
	rampColorB       colorful.Color
	scaleRamp        bool
}

func NewProgress(opts ...ProgressOption) *ProgressModel {
	m := ProgressModel{
		id:             nextID(),
		width:          defaultWidth,
		Full:           '█',
		FullColor:      lipgloss.Color("#7571F9"),
		Empty:          '░',
		EmptyColor:     lipgloss.Color("#606060"),
		ShowPercentage: true,
		PercentFormat:  " %3.0f%%",
	}

	for _, opt := range opts {
		opt(&m)
	}

	if !m.springCustomized {
		m.SetSpringOptions(defaultFrequency, defaultDamping)
	}

	return &m
}

func (m *ProgressModel) Init() tea.Cmd {
	return nil
}

func (m *ProgressModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case FrameMsg:
		if msg.id != m.id || msg.tag != m.tag {
			return m, nil
		}

		if !m.IsAnimating() {
			return m, nil
		}

		m.percentShown, m.velocity = m.spring.Update(m.percentShown, m.velocity, m.targetPercent)
		return m, m.nextFrame()

	default:
		return m, nil
	}
}

func (m *ProgressModel) SetSpringOptions(frequency, damping float64) {
	m.spring = harmonica.NewSpring(harmonica.FPS(fps), frequency, damping)
}

func (m ProgressModel) Percent() float64 {
	return m.targetPercent
}

func (m *ProgressModel) SetPercent(p float64) tea.Cmd {
	m.targetPercent = math.Max(0, math.Min(1, p))
	m.tag++
	return m.nextFrame()
}

func (m *ProgressModel) IncrPercent(v float64) tea.Cmd {
	return m.SetPercent(m.Percent() + v)
}

func (m *ProgressModel) DecrPercent(v float64) tea.Cmd {
	return m.SetPercent(m.Percent() - v)
}

func (m *ProgressModel) View() string {
	return m.ViewAs(m.percentShown)
}

func (m *ProgressModel) ViewAs(percent float64) string {
	b := strings.Builder{}
	percentView := m.percentageView(percent)
	m.barView(&b, percent, ansi.StringWidth(percentView))
	b.WriteString(percentView)
	return b.String()
}

func (m *ProgressModel) SetWidth(w int) {
	m.width = w
}

func (m ProgressModel) Width() int {
	return m.width
}

func (m *ProgressModel) nextFrame() tea.Cmd {
	return tea.Tick(time.Second/time.Duration(fps), func(time.Time) tea.Msg {
		return FrameMsg{id: m.id, tag: m.tag}
	})
}

func (m ProgressModel) barView(b *strings.Builder, percent float64, textWidth int) {
	var (
		tw = max(0, m.width-textWidth)
		fw = int(math.Round((float64(tw) * percent)))
		p  float64
	)

	fw = max(0, min(tw, fw))

	if m.useRamp {
		for i := 0; i < fw; i++ {
			if fw == 1 {
				p = 0.5
			} else if m.scaleRamp {
				p = float64(i) / float64(fw-1)
			} else {
				p = float64(i) / float64(tw-1)
			}
			c := m.rampColorA.BlendLuv(m.rampColorB, p)
			b.WriteString(lipgloss.NewStyle().Foreground(c).Render(string(m.Full)))
		}
	} else {
		b.WriteString(lipgloss.NewStyle().
			Foreground(m.FullColor).
			Render(strings.Repeat(string(m.Full), fw)))
	}

	n := max(0, tw-fw)
	b.WriteString(lipgloss.NewStyle().
		Foreground(m.EmptyColor).
		Render(strings.Repeat(string(m.Empty), n)))
}

func (m ProgressModel) percentageView(percent float64) string {
	if !m.ShowPercentage {
		return ""
	}
	percent = math.Max(0, math.Min(1, percent))
	percentage := fmt.Sprintf(m.PercentFormat, percent*100)
	percentage = m.PercentageStyle.Inline(true).Render(percentage)
	return percentage
}

func (m *ProgressModel) setRamp(colorA, colorB string, scaled bool) {
	a, _ := colorful.Hex(colorA)
	b, _ := colorful.Hex(colorB)

	m.useRamp = true
	m.scaleRamp = scaled
	m.rampColorA = a
	m.rampColorB = b
}

func (m *ProgressModel) IsAnimating() bool {
	dist := math.Abs(m.percentShown - m.targetPercent)
	return !(dist < 0.001 && m.velocity < 0.01)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
