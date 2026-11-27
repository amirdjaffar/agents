package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func catalog(t *testing.T) []Group {
	t.Helper()
	groups, err := loadCatalog(catalogYAML)
	if err != nil {
		t.Fatal(err)
	}
	return groups
}

func TestCatalogRejectsMissingOpposite(t *testing.T) {
	bad := "- {id: g, title: G, summary: s, parts: [{id: p, summary: s, rule: r}]}"
	if _, err := loadCatalog([]byte(bad)); err == nil {
		t.Error("a part without an opposite was accepted")
	}
}

func TestSelectionRoundTrip(t *testing.T) {
	groups := catalog(t)
	sel := defaults(groups)
	sel["code"]["tests"] = Opposite
	sel["code"]["comments"] = Unset
	sel.set(groups[0], "delegated", Rule)
	got, found := readSelection(render(groups, sel), groups)
	if !found {
		t.Fatal("marker not found")
	}
	if got.encode(groups) != sel.encode(groups) {
		t.Errorf("got %s\nwant %s", got.encode(groups), sel.encode(groups))
	}
}

func TestRender(t *testing.T) {
	groups := catalog(t)
	sel := defaults(groups)
	sel["code"]["tests"] = Opposite
	out := render(groups, sel)
	for _, want := range []string{"## Supervision: reviewed", "- Tests are optional.", "- Understanding comes before features."} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q", want)
		}
	}
	if strings.Contains(out, "## Planning") {
		t.Error("unset planning was rendered")
	}
}

func TestSplice(t *testing.T) {
	section := "<!-- agents a=b -->\nrules\n<!-- /agents -->\n"
	newer := "<!-- agents a=c -->\nnew rules\n<!-- /agents -->\n"
	doc := splice("# repo\n\nOwn text.", section)
	if want := "# repo\n\nOwn text.\n\n" + section; doc != want {
		t.Fatalf("append: got %q", doc)
	}
	doc += "\nMore own text.\n"
	if splice(doc, section) != doc {
		t.Error("rebuild with the same section changed the doc")
	}
	if want := "# repo\n\nOwn text.\n\n" + newer + "\nMore own text.\n"; splice(doc, newer) != want {
		t.Errorf("replace: got %q", splice(doc, newer))
	}
}

func key(s string) tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: rune(s[0]), Text: s}
}

func TestModelKeys(t *testing.T) {
	groups := catalog(t)
	var m tea.Model = newModel(groups, defaults(groups), "AGENTS.md", "")
	m, _ = m.Update(key("j"))
	m, _ = m.Update(key("r"))
	sel := m.(model).sel
	if sel["supervision"]["reviewed"] != Rule || len(sel["supervision"]) != 1 {
		t.Errorf("supervision = %v", sel["supervision"])
	}
	for range len(groups[0].Parts) - 1 {
		m, _ = m.Update(key("j"))
	}
	m, _ = m.Update(key("o"))
	if got := m.(model).sel["code"]["understanding"]; got != Opposite {
		t.Errorf("understanding = %v, want opposite", got)
	}
	m, _ = m.Update(key("w"))
	m, _ = m.Update(key("y"))
	if !m.(model).write {
		t.Error("w then y did not confirm")
	}
}

func TestWriteRules(t *testing.T) {
	dir := t.TempDir()
	created, err := writeRules(dir, "", false, "<!-- agents -->\n<!-- /agents -->\n")
	if err != nil || !created {
		t.Fatalf("created=%v err=%v", created, err)
	}
	doc, _ := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	if !strings.HasPrefix(string(doc), "# "+filepath.Base(dir)+"\n\n<!-- agents") {
		t.Errorf("doc = %q", doc)
	}
	if created, _ := writeRules(dir, string(doc), true, "<!-- agents -->\n<!-- /agents -->\n"); created {
		t.Error("CLAUDE.md created twice")
	}
}
