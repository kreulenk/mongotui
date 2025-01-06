// MIT License
//
// Copyright (c) 2020-2023 Charmbracelet, Inc
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// This table is a modified version of the bubbles table component which can be found at:
// https://github.com/charmbracelet/bubbles/tree/master/table

package table

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
	"mtui/pkg/mongodata"
)

type cursorColumn int

const (
	databasesColumn cursorColumn = iota
	collectionsColumn
	dataColumn
)

// Model defines a state for the table widget.
type Model struct {
	KeyMap KeyMap
	Help   help.Model

	cols            []Column
	rows            []Row
	cursorColumn    cursorColumn
	cursorRow       int
	selectedDb      string // TODO Potentially refactor both of these selected fields into one field
	selectedDbIndex int    // To remember what database we are on when we go left back from collections row
	focus           bool
	styles          Styles

	viewport viewport.Model
	start    int
	end      int

	engine *mongodata.Engine
}

// Row represents one line in the table.
type Row []string

// Column defines the table structure.
type Column struct {
	Title string
	Width int
}

// KeyMap defines keybindings. It satisfies to the help.KeyMap interface, which
// is used to render the help menu.
type KeyMap struct {
	LineUp       key.Binding
	LineDown     key.Binding
	PageUp       key.Binding
	PageDown     key.Binding
	HalfPageUp   key.Binding
	HalfPageDown key.Binding
	GotoTop      key.Binding
	GotoBottom   key.Binding
	Right        key.Binding
	Left         key.Binding
}

// ShortHelp implements the KeyMap interface.
func (km KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{km.LineUp, km.LineDown, km.Right, km.Left}
}

// FullHelp implements the KeyMap interface.
func (km KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{km.Right, km.Left},
		{km.LineUp, km.LineDown, km.GotoTop, km.GotoBottom},
		{km.PageUp, km.PageDown, km.HalfPageUp, km.HalfPageDown},
	}
}

// DefaultKeyMap returns a default set of keybindings.
func DefaultKeyMap() KeyMap {
	const spacebar = " "
	return KeyMap{
		LineUp: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		LineDown: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("b", "pgup"),
			key.WithHelp("b/pgup", "page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("f", "pgdown", spacebar),
			key.WithHelp("f/pgdn", "page down"),
		),
		HalfPageUp: key.NewBinding(
			key.WithKeys("u", "ctrl+u"),
			key.WithHelp("u", "½ page up"),
		),
		HalfPageDown: key.NewBinding(
			key.WithKeys("d", "ctrl+d"),
			key.WithHelp("d", "½ page down"),
		),
		GotoTop: key.NewBinding(
			key.WithKeys("home", "g"),
			key.WithHelp("g/home", "go to start"),
		),
		GotoBottom: key.NewBinding(
			key.WithKeys("end", "G"),
			key.WithHelp("G/end", "go to end"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", "right"),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "left"),
		),
	}
}

// Styles contains style definitions for this list component. By default, these
// values are generated by DefaultStyles.
type Styles struct {
	Header   lipgloss.Style
	Cell     lipgloss.Style
	Selected lipgloss.Style
}

// DefaultStyles returns a set of default style definitions for this table.
func DefaultStyles() Styles {
	return Styles{
		Selected: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212")),
		Header:   lipgloss.NewStyle().Bold(true).Padding(0, 1),
		Cell:     lipgloss.NewStyle().Padding(0, 1),
	}
}

// SetStyles sets the table styles.
func (m *Model) SetStyles(s Styles) {
	m.styles = s
	m.UpdateViewport()
}

// Option is used to set options in New. For example:
//
//	table := New(WithColumns([]Column{{Title: "ID", Width: 10}}))
type Option func(*Model)

// New creates a new baseModel for the table widget.
func New(engine *mongodata.Engine, opts ...Option) Model {
	databases := getSortedDatabasesByName(engine.Server.Databases)

	m := Model{
		cursorRow:       0,
		cursorColumn:    databasesColumn,
		viewport:        viewport.New(0, 20),
		selectedDb:      databases[0],
		selectedDbIndex: 0,

		cols: []Column{
			{Title: "Databases"},
			{Title: "Collections"},
		},
		rows: []Row{databases},

		KeyMap: DefaultKeyMap(),
		Help:   help.New(),
		styles: DefaultStyles(),

		engine: engine,
	}

	for _, opt := range opts {
		opt(&m)
	}

	m.updateTableRows()
	m.UpdateViewport()

	return m
}

// WithHeight sets the height of the table.
func WithHeight(h int) Option {
	return func(m *Model) {
		m.viewport.Height = h - lipgloss.Height(m.headersView())
	}
}

// WithWidth sets the width of the table.
func WithWidth(w int) Option {
	return func(m *Model) {
		m.viewport.Width = w
	}
}

// WithFocused sets the focus state of the table.
func WithFocused(f bool) Option {
	return func(m *Model) {
		m.focus = f
	}
}

// WithStyles sets the table styles.
func WithStyles(s Styles) Option {
	return func(m *Model) {
		m.styles = s
	}
}

// WithKeyMap sets the key map.
func WithKeyMap(km KeyMap) Option {
	return func(m *Model) {
		m.KeyMap = km
	}
}

// Update is the Bubble Tea update loop.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.focus {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.KeyMap.LineUp):
			m.MoveUp(1)
		case key.Matches(msg, m.KeyMap.LineDown):
			m.MoveDown(1)
		case key.Matches(msg, m.KeyMap.PageUp):
			m.MoveUp(m.viewport.Height)
		case key.Matches(msg, m.KeyMap.PageDown):
			m.MoveDown(m.viewport.Height)
		case key.Matches(msg, m.KeyMap.HalfPageUp):
			m.MoveUp(m.viewport.Height / 2)
		case key.Matches(msg, m.KeyMap.HalfPageDown):
			m.MoveDown(m.viewport.Height / 2)
		case key.Matches(msg, m.KeyMap.LineDown):
			m.MoveDown(1)
		case key.Matches(msg, m.KeyMap.GotoTop):
			m.GotoTop()
		case key.Matches(msg, m.KeyMap.GotoBottom):
			m.GotoBottom()
		case key.Matches(msg, m.KeyMap.Right):
			m.GoRight()
		case key.Matches(msg, m.KeyMap.Left):
			m.GoLeft()
		}
	}

	return m, nil
}

// Focused returns the focus state of the table.
func (m Model) Focused() bool {
	return m.focus
}

// Focus focuses the table, allowing the user to move around the rows and
// interact.
func (m *Model) Focus() {
	m.focus = true
	m.UpdateViewport()
}

// Blur blurs the table, preventing selection or movement.
func (m *Model) Blur() {
	m.focus = false
	m.UpdateViewport()
}

// View renders the component.
func (m Model) View() string {
	return m.headersView() + "\n" + m.viewport.View()
}

// HelpView is a helper method for rendering the help menu from the keymap.
// Note that this view is not rendered by default and you must call it
// manually in your application, where applicable.
func (m Model) HelpView() string {
	return m.Help.View(m.KeyMap)
}

// UpdateViewport updates the list content based on the previously defined
// columns and rows.
func (m *Model) UpdateViewport() {
	renderedRows := make([]string, 0, len(m.rows))

	// Render only rows from: m.cursorRow-m.viewport.Height to: m.cursorRow+m.viewport.Height
	// Constant runtime, independent of number of rows in a table.
	// Limits the number of renderedRows to a maximum of 2*m.viewport.Height
	if m.cursorRow >= 0 {
		m.start = clamp(m.cursorRow-m.viewport.Height, 0, m.cursorRow)
	} else {
		m.start = 0
	}
	m.end = clamp(m.cursorRow+m.viewport.Height, m.cursorRow, len(m.rows))
	for i := m.start; i < m.end; i++ {
		renderedRows = append(renderedRows, m.renderRow(i))
	}

	m.viewport.SetContent(
		lipgloss.JoinVertical(lipgloss.Left, renderedRows...),
	)
}

// SelectedRow returns the selected row.
func (m Model) SelectedRow() Row {
	if m.cursorRow < 0 || m.cursorRow >= len(m.rows) {
		return nil
	}

	return m.rows[m.cursorRow]
}

// SelectedCell returns the text within the currently highlighted cell
func (m Model) SelectedCell() string {
	if m.cursorRow < 0 || m.cursorRow >= len(m.rows) {
		return ""
	}

	return m.rows[m.cursorRow][m.cursorColumn]
}

// Rows returns the current rows.
func (m Model) Rows() []Row {
	return m.rows
}

// Columns returns the current columns.
func (m Model) Columns() []Column {
	return m.cols
}

// SetRows sets a new rows state.
func (m *Model) SetRows(r []Row) {
	m.rows = r
	m.UpdateViewport()
}

// SetColumns sets a new columns state.
func (m *Model) SetColumns(c []Column) {
	m.cols = c
	m.UpdateViewport()
}

// SetWidth sets the width of the viewport of the table.
func (m *Model) SetWidth(w int) {
	m.viewport.Width = w
	m.UpdateViewport()
}

// SetHeight sets the height of the viewport of the table.
func (m *Model) SetHeight(h int) {
	m.viewport.Height = h - lipgloss.Height(m.headersView())
	m.UpdateViewport()
}

// Height returns the viewport height of the table.
func (m Model) Height() int {
	return m.viewport.Height
}

// Width returns the viewport width of the table.
func (m Model) Width() int {
	return m.viewport.Width
}

// Cursor returns the index of the selected row.
func (m Model) Cursor() int {
	return m.cursorRow
}

// SetCursor sets the cursorRow position in the table.
func (m *Model) SetCursor(n int) {
	m.cursorRow = clamp(n, 0, len(m.rows)-1)
	m.UpdateViewport()
}

// MoveUp moves the selection up by any number of rows.
// It can not go above the first row.
func (m *Model) MoveUp(n int) {
	m.cursorRow = max(0, m.cursorRow-1)
	if m.cursorColumn == databasesColumn {
		m.selectedDbIndex = m.cursorRow
		m.selectedDb = m.SelectedCell()
	}

	switch {
	case m.start == 0:
		m.viewport.SetYOffset(clamp(m.viewport.YOffset, 0, m.cursorRow))
	case m.start < m.viewport.Height:
		m.viewport.YOffset = (clamp(clamp(m.viewport.YOffset+n, 0, m.cursorRow), 0, m.viewport.Height))
	case m.viewport.YOffset >= 1:
		m.viewport.YOffset = clamp(m.viewport.YOffset+n, 1, m.viewport.Height)
	}
	m.updateTableRows()
	m.UpdateViewport()
}

// MoveDown moves the selection down by any number of rows.
// It can not go below the last row.
func (m *Model) MoveDown(n int) {
	if m.cursorColumn == databasesColumn {
		m.cursorRow = clamp(m.cursorRow+n, 0, len(m.engine.Server.Databases)-1)
		m.selectedDbIndex = m.cursorRow
		m.selectedDb = m.SelectedCell()
	} else { // collections column
		m.cursorRow = clamp(m.cursorRow+n, 0, len(m.engine.Server.Databases[m.selectedDb].Collections)-1)
	}

	m.updateTableRows()
	m.UpdateViewport()

	switch {
	case m.end == len(m.rows) && m.viewport.YOffset > 0:
		m.viewport.SetYOffset(clamp(m.viewport.YOffset-n, 1, m.viewport.Height))
	case m.cursorRow > (m.end-m.start)/2 && m.viewport.YOffset > 0:
		m.viewport.SetYOffset(clamp(m.viewport.YOffset-n, 1, m.cursorRow))
	case m.viewport.YOffset > 1:
	case m.cursorRow > m.viewport.YOffset+m.viewport.Height-1:
		m.viewport.SetYOffset(clamp(m.viewport.YOffset+1, 0, 1))
	}
}

// MoveRight moves the column to the right.
func (m *Model) MoveRight() {
	if m.cursorColumn == databasesColumn {
		m.selectedDb = m.SelectedCell()
		m.cursorColumn = collectionsColumn
		m.selectedDbIndex = m.cursorRow
		m.cursorRow = 0
	}
	m.updateTableRows()
	m.UpdateViewport()
}

// MoveLeft moves the column to the left.
func (m *Model) MoveLeft() {
	if m.cursorColumn == collectionsColumn {
		m.cursorColumn = databasesColumn
		m.cursorRow = m.selectedDbIndex
		m.selectedDb = m.SelectedCell()
	}
	m.updateTableRows()
	m.UpdateViewport()
}

// GotoTop moves the selection to the first row.
func (m *Model) GotoTop() {
	m.MoveUp(m.cursorRow)
}

// GotoBottom moves the selection to the last row.
func (m *Model) GotoBottom() {
	m.MoveDown(len(m.rows))
}

// GoRight moves to the next column.
func (m *Model) GoRight() {
	m.MoveRight()
}

func (m *Model) GoLeft() {
	m.MoveLeft()
}

func (m Model) headersView() string {
	s := make([]string, 0, len(m.cols))
	for _, col := range m.cols {
		if col.Width <= 0 {
			continue
		}
		style := lipgloss.NewStyle().Width(col.Width).MaxWidth(col.Width).Inline(true)
		renderedCell := style.Render(runewidth.Truncate(col.Title, col.Width, "…"))
		s = append(s, m.styles.Header.Render(renderedCell))
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, s...)
}

func (m *Model) renderRow(r int) string {
	s := make([]string, 0, len(m.cols))
	for i, value := range m.rows[r] {
		if m.cols[i].Width <= 0 {
			continue
		}
		style := lipgloss.NewStyle().Width(m.cols[i].Width).MaxWidth(m.cols[i].Width).Inline(true)
		renderedCell := m.styles.Cell.Render(style.Render(runewidth.Truncate(value, m.cols[i].Width, "…")))
		if (m.cursorColumn == databasesColumn && cursorColumn(i) == databasesColumn && r == m.cursorRow) ||
			(m.cursorColumn == collectionsColumn && cursorColumn(i) == databasesColumn && r == m.selectedDbIndex) ||
			(m.cursorColumn == collectionsColumn && cursorColumn(i) == collectionsColumn && r == m.cursorRow) {
			renderedCell = m.styles.Selected.Render(renderedCell)
		}
		s = append(s, renderedCell)
	}

	row := lipgloss.JoinHorizontal(lipgloss.Top, s...)
	return row
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

func clamp(v, low, high int) int {
	return min(max(v, low), high)
}
