package main

import (
	"os/exec"
	"strings"
)

func getRecentHeads() ([]string, error) {
	out, err := exec.Command("git", "for-each-ref", "--sort=-committerdate", "--count=50", "refs/heads/").Output()
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	refs := make([]string, 0, len(lines))
	for _, line := range lines {
		parts := strings.Split(line, "\t")
		refs = append(refs, parts[1])
	}
	return refs, nil
}

func getLogs(ref string) ([]string, error) {
	out, err := exec.Command("git", "log", `--pretty=format:"%H %s"`, "-n", "20", ref).Output()
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	return lines, nil
}
