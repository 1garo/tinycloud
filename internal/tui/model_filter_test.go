package tui

import (
	"testing"

	"github.com/1garo/tinycloud/internal/store"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

func TestFilteringDoesNotTriggerShortcuts(t *testing.T) {
	m := NewModel(store.New(t.TempDir()+"/state.json"), nil).(*model)
	m.list.SetFilterState(list.Filtering)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	current := updated.(*model)
	if current.busy {
		t.Fatal("typing d while filtering should not start a deployment")
	}
}
