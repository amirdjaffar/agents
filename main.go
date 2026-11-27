package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	tea "charm.land/bubbletea/v2"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "agents:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) > 1 {
		return errors.New("usage: agents [repo]")
	}
	dir := "."
	if len(args) == 1 {
		dir = args[0]
	}
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		return fmt.Errorf("no such directory: %s", dir)
	}
	groups, err := loadCatalog(catalogYAML)
	if err != nil {
		return err
	}

	file := filepath.Join(dir, "AGENTS.md")
	data, err := os.ReadFile(file)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	exists := err == nil
	sel, found := readSelection(string(data), groups)
	status := "New file. Your choices become its rules section."
	switch {
	case found:
		status = "Has a rules section. Its choices are loaded; text outside it is kept."
	case exists:
		status = "Has no rules section yet. One will be added at the end; existing text is kept."
	}

	final, err := tea.NewProgram(newModel(groups, sel, file, status)).Run()
	if err != nil {
		return err
	}
	if m := final.(model); !m.write {
		fmt.Println("Nothing written.")
		return nil
	}
	created, err := writeRules(dir, string(data), exists, render(groups, sel))
	if err != nil {
		return err
	}
	fmt.Println("Wrote", file)
	if created {
		fmt.Println("Created", filepath.Join(dir, "CLAUDE.md")+", which points Claude Code to AGENTS.md")
	}
	return nil
}

// writeRules splices section into AGENTS.md and adds a CLAUDE.md if there is
// none, reporting whether it did.
func writeRules(dir, doc string, exists bool, section string) (bool, error) {
	if !exists {
		abs, err := filepath.Abs(dir)
		if err != nil {
			return false, err
		}
		doc = "# " + filepath.Base(abs) + "\n"
	}
	if err := os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte(splice(doc, section)), 0o644); err != nil {
		return false, err
	}
	claude := filepath.Join(dir, "CLAUDE.md")
	if _, err := os.Stat(claude); !errors.Is(err, fs.ErrNotExist) {
		return false, err
	}
	return true, os.WriteFile(claude, []byte("@AGENTS.md\n"), 0o644)
}
