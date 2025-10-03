package ascii

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss/v2"
	"github.com/charmbracelet/x/exp/charmtone"
)

//go:embed logo.ascii
var art string

func PrintLogo() string {
	var (
		b      strings.Builder
		lines  = strings.Split(art, "\n")
		colors = []string{"#24FFFC", "#1AD9D6", "#10B1AE", "#068986"}
		step   = len(lines) / len(colors)
	)

	for i, l := range lines {
		n := clamp(0, len(colors)-1, i/step)
		b.WriteString(colorize(colors[n], l))
		b.WriteRune('\n')
	}

	return b.String()
}

func PrintSpecs() string {
	var b strings.Builder

	titleStyle := lipgloss.NewStyle().
		Foreground(charmtone.Mustard).
		Bold(true)

	keyStyle := lipgloss.NewStyle().
		Foreground(charmtone.Zinc)

	valueStyle := lipgloss.NewStyle().
		Foreground(charmtone.Butter)

	printInfo := func(key, value string) {
		fmt.Fprintln(&b, keyStyle.Render(key), valueStyle.Render(value))
	}

	b.WriteString("\n")
	b.WriteString(titleStyle.Render("Cerebellum-ITM/Catalyst") + "\n")
	b.WriteString("\n")

	printInfo("Description", "TUI for the RuneCraft Backend")
	printInfo("Framework", "Charmbracelet (Bubble Tea, Lipgloss)")
	printInfo("Language", "Go (from go.mod)")
	printInfo("Architecture", "Model-View-Update (MVU)")
	printInfo("Backend", "RuneCraft (via SSH)")
	printInfo("API Format", "JSON over stdout")
	printInfo("Config", "~/.config/Catalyst/config.toml")
	printInfo("History", "SQLite DB")
	printInfo("Caching", "In-memory Spellbook")

	// --- Block Colors ---
	var darkColorBlocks []string
	var colorBlocks []string
	var lightColorBlocks []string
	var grayColorBlocks []string

	for i := charmtone.Cumin; i < charmtone.Pepper; i++ {
		k := charmtone.Keys()[i]
		style := lipgloss.NewStyle().
			SetString("   ").
			Background(k)
		if i%3 == 0 {
			darkColorBlocks = append(darkColorBlocks, style.String())
		}
		if i%3 == 1 {
			colorBlocks = append(colorBlocks, style.String())
		}
		if i%3 == 2 {
			lightColorBlocks = append(lightColorBlocks, style.String())
		}
	}

	for i := charmtone.Pepper; i <= charmtone.Butter; i++ {
		k := charmtone.Keys()[i]
		style := lipgloss.NewStyle().
			SetString("   ").
			Background(k)
		grayColorBlocks = append(grayColorBlocks, style.String())
	}

	b.WriteString("\n")
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, grayColorBlocks...) + "\n")
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, darkColorBlocks...) + "\n")
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, colorBlocks...) + "\n")
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, lightColorBlocks...) + "\n")

	return b.String()
}

func colorize(c, s string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(c)).Render(s)
}

func clamp(v, low, high int) int {
	if high < low {
		low, high = high, low
	}
	return min(high, max(low, v))
}
