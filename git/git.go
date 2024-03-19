package git

import (
	"fmt"
	"os/exec"
	"strings"
)

func GetRecentHeads() ([]string, error) {
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

type Commit struct {
	Hash    string
	Message string
}

func GetLogs(ref string) ([]Commit, error) {
	out, err := exec.Command("git", "log", `--pretty=format:%H	%s`, "-n", "20", ref).Output()
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	commits := make([]Commit, 0, len(lines))
	for _, line := range lines {
		parts := strings.Split(line, "\t")
		if len(parts) != 2 {
			return nil, fmt.Errorf("unexpected log line: \"%s\"", line)
		}
		commits = append(commits, Commit{parts[0], parts[1]})

	}
	return commits, nil
}
