package main

import (
	"fmt"
	"log"
	"os/exec"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("212")).
			Bold(true)

	outputStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86"))

	subtleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	activeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("212")).
			Bold(true)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(1, 2)

	mainStyle = lipgloss.NewStyle().
			MarginLeft(2).
			MarginTop(1)
)

type model struct {
	Choice   int
	Chosen   bool
	Quitting bool
	Output   string
	Width    int
	Height   int
}

var menuItems = []string{
	"Running Containers",
	"All Containers",
	"Images",
	"Volumes",
	"Networks",
	"Docker Version",
	"Disk Usage",
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height

	case tea.KeyPressMsg:

		switch msg.String() {

		case "q", "esc", "ctrl+c":
			m.Quitting = true
			return m, tea.Quit
		}
	}

	if !m.Chosen {
		return updateChoices(msg, m)
	}

	return updateChosen(msg, m)
}

func updateChoices(msg tea.Msg, m model) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {

	case tea.KeyPressMsg:

		switch msg.String() {

		case "j", "down":
			if m.Choice < len(menuItems)-1 {
				m.Choice++
			}

		case "k", "up":
			if m.Choice > 0 {
				m.Choice--
			}

		case "g":
			m.Choice = 0

		case "G":
			m.Choice = len(menuItems) - 1

		case "enter":

			m.Chosen = true

			// run only once
			m.Output = runDockerCommand(m.Choice)

			return m, nil
		}
	}

	return m, nil
}

func updateChosen(msg tea.Msg, m model) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {

	case tea.KeyPressMsg:

		switch msg.String() {

		case "b", "enter":

			m.Chosen = false
			m.Output = ""

			return m, nil

		case "r":

			// manual refresh only
			m.Output = runDockerCommand(m.Choice)

			return m, nil
		}
	}

	return m, nil
}

func runDockerCommand(choice int) string {

	var cmd *exec.Cmd

	switch choice {

	case 0:
		cmd = exec.Command("docker", "ps")

	case 1:
		cmd = exec.Command("docker", "ps", "-a")

	case 2:
		cmd = exec.Command("docker", "images")

	case 3:
		cmd = exec.Command("docker", "volume", "ls")

	case 4:
		cmd = exec.Command("docker", "network", "ls")

	case 5:
		cmd = exec.Command("docker", "--version")

	case 6:
		cmd = exec.Command("docker", "system", "df")

	default:
		return "Unknown option"
	}

	out, err := cmd.CombinedOutput()

	if err != nil {
		log.Println(err)
		return err.Error()
	}

	return strings.TrimSpace(string(out))
}

func (m model) View() tea.View {

	if m.Quitting {
		return tea.NewView("\n Goodbye 👋\n")
	}

	var content string

	if !m.Chosen {
		content = choicesView(m)
	} else {
		content = chosenView(m)
	}

	return tea.NewView(
		mainStyle.Render(content),
	)
}

func choicesView(m model) string {

	var b strings.Builder

	b.WriteString(
		titleStyle.Render("🐳 Docker Dashboard"),
	)

	b.WriteString("\n\n")

	for i, item := range menuItems {

		cursor := "[ ]"

		if i == m.Choice {
			cursor = activeStyle.Render("[x]")
			item = activeStyle.Render(item)
		}

		b.WriteString(
			fmt.Sprintf("%s %s\n", cursor, item),
		)
	}

	b.WriteString("\n")

	b.WriteString(
		subtleStyle.Render(
			"j/k • ↑/↓ navigate • enter select • g top • G bottom • q quit",
		),
	)

	return boxStyle.Render(b.String())
}

func chosenView(m model) string {

	title := titleStyle.Render(
		fmt.Sprintf("📄 %s", menuItems[m.Choice]),
	)

	footer := subtleStyle.Render(
		"\n\nb/enter back • r refresh • q quit",
	)

	view := fmt.Sprintf(
		"%s\n\n%s\n\n%s",
		title,
		outputStyle.Render(m.Output),
		footer,
	)

	width := m.Width - 6

	if width < 20 {
		width = 20
	}

	return boxStyle.Width(width).Render(view)
}

func main() {

	initialModel := model{
		Choice:   0,
		Chosen:   false,
		Quitting: false,
		Output:   "",
	}

	p := tea.NewProgram(initialModel)

	if _, err := p.Run(); err != nil {
		fmt.Println("error:", err)
	}
}
