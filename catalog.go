package main

import (
	_ "embed"
	"fmt"

	"go.yaml.in/yaml/v3"
)

//go:embed catalog.yaml
var catalogYAML []byte

type Group struct {
	ID      string `yaml:"id"`
	Title   string `yaml:"title"`
	Summary string `yaml:"summary"`
	One     bool   `yaml:"one"`
	Parts   []Part `yaml:"parts"`
}

// A Part has a rule and its opposite. Parts of a group that picks one have
// only a rule.
type Part struct {
	ID       string `yaml:"id"`
	Summary  string `yaml:"summary"`
	Default  string `yaml:"default"`
	Rule     string `yaml:"rule"`
	Opposite string `yaml:"opposite"`
}

func loadCatalog(data []byte) ([]Group, error) {
	var groups []Group
	if err := yaml.Unmarshal(data, &groups); err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	for _, g := range groups {
		if g.ID == "" || g.Title == "" || g.Summary == "" || len(g.Parts) == 0 || seen[g.ID] {
			return nil, fmt.Errorf("group %q: needs a unique id, title, summary and parts", g.ID)
		}
		seen[g.ID] = true
		if err := checkParts(g); err != nil {
			return nil, fmt.Errorf("group %s: %w", g.ID, err)
		}
	}
	return groups, nil
}

func checkParts(g Group) error {
	seen := map[string]bool{}
	defaults := 0
	for _, p := range g.Parts {
		if p.ID == "" || p.Summary == "" || p.Rule == "" || seen[p.ID] {
			return fmt.Errorf("part %q: needs a unique id, summary and rule", p.ID)
		}
		seen[p.ID] = true
		if g.One != (p.Opposite == "") {
			return fmt.Errorf("part %s: only parts of a group that picks many have an opposite", p.ID)
		}
		switch p.Default {
		case "":
		case "rule":
			defaults++
		case "opposite":
			if g.One {
				return fmt.Errorf("part %s: default must be rule", p.ID)
			}
		default:
			return fmt.Errorf("part %s: default must be rule or opposite", p.ID)
		}
	}
	if g.One && defaults > 1 {
		return fmt.Errorf("more than one default")
	}
	return nil
}
