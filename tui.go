package main

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

var (
	bold     = lipgloss.NewStyle().Bold(true)
	faint    = lipgloss.NewStyle().Faint(true)
	green    = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	red      = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	selected = lipgloss.NewStyle().Reverse(true)
)

const (
	help = "↑↓ move · tab group · ←→ change · r/o/u rule/opposite/unset · w write · q quit"
	// Lines around the list: file, status, a blank line and the "more above"
	// line; the "more below" line, a blank line, the divider, a blank line
	// and help, plus the detail.
	chrome = 4 + 5
	// Lines kept visible above and below the cursor when scrolling.
	margin = 2
)

// A row is a group heading (part < 0) or a part.
type row struct{ group, part int }

type model struct {
	groups       []Group
	sel          Selection
	rows         []row
	cursor       int
	offset       int
	file, status string
	width        int
	height       int
	detailHeight int
	confirming   bool
	write        bool
}

func newModel(groups []Group, sel Selection, file, status string) model {
	m := model{groups: groups, sel: sel, file: file, status: status, cursor: 1}
	for gi, g := range groups {
		m.rows = append(m.rows, row{gi, -1})
		for pi := range g.Parts {
			m.rows = append(m.rows, row{gi, pi})
		}
	}
	m.detailHeight = m.tallestDetail()
	return m
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.detailHeight = min(m.tallestDetail(), max(m.height/4, 4))
	case tea.KeyPressMsg:
		if m.confirming {
			if k := msg.String(); k == "y" || k == "enter" {
				m.write = true
				return m, tea.Quit
			}
			m.confirming = false
			return m, nil
		}
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		case "up", "k":
			m.move(-1)
		case "down", "j":
			m.move(1)
		case "tab":
			m.nextGroup()
		case "space", "enter", "right", "l":
			m.cycle(1)
		case "left", "h":
			m.cycle(-1)
		case "r":
			m.setState(Rule)
		case "o":
			m.setState(Opposite)
		case "u", "backspace":
			m.setState(Unset)
		case "w":
			m.confirming = true
		}
	}
	m.scroll()
	return m, nil
}

func (m *model) move(d int) {
	for i := m.cursor + d; i >= 0 && i < len(m.rows); i += d {
		if m.rows[i].part >= 0 {
			m.cursor = i
			return
		}
	}
}

// nextGroup moves to the first part of the next group, wrapping around.
func (m *model) nextGroup() {
	next := (m.rows[m.cursor].group + 1) % len(m.groups)
	for i, r := range m.rows {
		if r.group == next && r.part == 0 {
			m.cursor = i
			return
		}
	}
}

func (m *model) focused() (Group, Part) {
	r := m.rows[m.cursor]
	g := m.groups[r.group]
	return g, g.Parts[r.part]
}

func (m *model) state() State {
	g, p := m.focused()
	return m.sel[g.ID][p.ID]
}

// cycle steps through unset, rule and opposite. Parts of a group that picks
// one only toggle between unset and rule.
func (m *model) cycle(d int) {
	g, _ := m.focused()
	if g.One {
		if m.state() == Rule {
			m.setState(Unset)
		} else {
			m.setState(Rule)
		}
		return
	}
	m.setState((m.state() + State(d) + 3) % 3)
}

func (m *model) setState(st State) {
	g, p := m.focused()
	if g.One && st == Opposite {
		return
	}
	m.sel.set(g, p.ID, st)
}

// line returns the line of row i in the list.
func (m *model) line(i int) int {
	n := i
	for _, r := range m.rows[1 : i+1] {
		if r.part < 0 {
			n++ // blank line before each heading but the first
		}
	}
	return n
}

func (m *model) listRoom() int {
	if m.height == 0 {
		return m.line(len(m.rows)-1) + 1
	}
	return max(m.height-chrome-m.detailHeight, 1)
}

// scroll moves the window over the list only when the cursor nears its edge.
func (m *model) scroll() {
	room, total := m.listRoom(), m.line(len(m.rows)-1)+1
	at := m.line(m.cursor)
	if at-margin < m.offset {
		m.offset = at - margin
	}
	if at+margin >= m.offset+room {
		m.offset = at + margin - room + 1
	}
	m.offset = max(min(m.offset, total-room), 0)
}

func (m model) View() tea.View {
	var list []string
	for i, r := range m.rows {
		g := m.groups[r.group]
		if r.part < 0 {
			if i > 0 {
				list = append(list, "")
			}
			list = append(list, bold.Render(g.Title)+faint.Render(" · "+g.Summary))
			continue
		}
		p := g.Parts[r.part]
		id := fmt.Sprintf("%-14s", p.ID)
		if i == m.cursor {
			id = selected.Render(id)
		}
		list = append(list, fmt.Sprintf("  %s %s %s", mark(g, m.sel[g.ID][p.ID]), id, faint.Render(p.Summary)))
	}
	end := min(m.offset+m.listRoom(), len(list))
	above, below := m.hiddenParts(end)
	list = list[m.offset:end]

	footer := faint.Render(help)
	if m.confirming {
		footer = bold.Render("Write the rules section to " + m.file + "? y/n")
	}
	lines := []string{bold.Render("agents") + " " + m.file, faint.Render(m.status), "", more("↑", above)}
	lines = append(lines, list...)
	lines = append(lines, more("↓", below), "", m.divider(), m.detail(m.cursor), "", footer)
	v := tea.NewView(strings.Join(lines, "\n"))
	v.AltScreen = true
	return v
}

// hiddenParts counts the parts scrolled out of view above and below a window
// of the list that ends at line end.
func (m model) hiddenParts(end int) (above, below int) {
	for i, r := range m.rows {
		switch at := m.line(i); {
		case r.part < 0:
		case at < m.offset:
			above++
		case at >= end:
			below++
		}
	}
	return above, below
}

func more(arrow string, n int) string {
	if n == 0 {
		return ""
	}
	return faint.Render(fmt.Sprintf("  %s %d more", arrow, n))
}

// divider separates the list from the detail and names the focused part.
func (m model) divider() string {
	g, p := m.focused()
	label := "── " + g.Title + " › " + p.ID + " "
	return faint.Render(label + strings.Repeat("─", max(m.width-lipgloss.Width(label), 0)))
}

func mark(g Group, st State) string {
	switch {
	case g.One && st == Rule:
		return green.Render("●")
	case g.One:
		return faint.Render("○")
	case st == Rule:
		return green.Render("✓")
	case st == Opposite:
		return red.Render("✗")
	}
	return faint.Render("·")
}

// detail shows what the part at row i writes in each state. It is clipped or
// padded to detailHeight, so the list above never moves.
func (m model) detail(i int) string {
	r := m.rows[i]
	g := m.groups[r.group]
	p := g.Parts[r.part]
	st := m.sel[g.ID][p.ID]
	text := lipgloss.NewStyle()
	if m.width > 12 {
		text = text.Width(m.width - 12)
	}
	show := func(s State, label, body string) string {
		line := lipgloss.JoinHorizontal(lipgloss.Top, fmt.Sprintf("%-12s", label), text.Render(body))
		if s != st {
			return faint.Render(line)
		}
		return line
	}
	out := show(Rule, "● writes", p.Rule)
	if !g.One {
		out = show(Rule, "✓ rule", p.Rule) + "\n" + show(Opposite, "✗ opposite", p.Opposite)
	}
	if m.detailHeight == 0 {
		return out
	}
	if lines := strings.Split(out, "\n"); len(lines) > m.detailHeight {
		out = strings.Join(lines[:m.detailHeight-1], "\n") + "\n" + faint.Render("…")
	}
	return lipgloss.NewStyle().Height(m.detailHeight).Render(out)
}

// tallestDetail returns the height of the tallest unclipped detail.
func (m model) tallestDetail() int {
	m.detailHeight = 0
	tallest := 0
	for i, r := range m.rows {
		if r.part >= 0 {
			tallest = max(tallest, lipgloss.Height(m.detail(i)))
		}
	}
	return tallest
}
