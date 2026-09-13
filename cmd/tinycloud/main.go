package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/1garo/tinycloud/internal/runtime"
	"github.com/1garo/tinycloud/internal/store"
	"github.com/1garo/tinycloud/internal/tui"
	"github.com/charmbracelet/bubbletea"
)

func main() {
	statePath := filepath.Join(".tinycloud", "state.json")
	appStore := store.New(statePath)
	if err := appStore.Load(); err != nil {
		fmt.Fprintf(os.Stderr, "load state: %v\n", err)
		os.Exit(1)
	}

	program := tea.NewProgram(tui.NewModel(appStore, runtime.NewDocker()))
	if _, err := program.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "run TinyCloud: %v\n", err)
		os.Exit(1)
	}
}
