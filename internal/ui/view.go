package ui

import "fmt"

func (m Model) View() string {
	if m.quitting {
		return ""
	}

	header := HeaderStyle.Render("netmon - Network Monitor")
	table := m.table.View()
	footer := FooterStyle.Render("Press q to quit")

	return fmt.Sprintf("%s\n%s\n%s\n", header, table, footer)
}
