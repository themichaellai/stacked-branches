package git

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

var refHeadRegexp = regexp.MustCompile(`^refs/heads/(.+)$`)

func GetRecentHeads() ([]string, error) {
	out, err := exec.Command("git", "for-each-ref", "--sort=-committerdate", "--count=50", "refs/heads/").Output()
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	refs := make([]string, 0, len(lines))
	for _, line := range lines {
		parts := strings.Split(line, "\t")
		nameMatch := refHeadRegexp.FindStringSubmatch(parts[1])
		if nameMatch == nil {
			return nil, fmt.Errorf("unexpected ref name: \"%s\"", parts[1])
		}
		refs = append(refs, nameMatch[1])
	}
	return refs, nil
}

type Commit struct {
	Hash    string
	Message string
}

const mainBranch = "origin/main"

func GetLogs(ref string) ([]Commit, error) {
	out, err := exec.Command("git", "log", `--pretty=format:%H	%s`, fmt.Sprintf("%s..%s", mainBranch, ref)).Output()
	if err != nil {
		return nil, err
	}
	trimmedOutput := strings.TrimSpace(string(out))
	if trimmedOutput == "" {
		return []Commit{}, nil
	}
	lines := strings.Split(trimmedOutput, "\n")
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

// Returns true if refA is an ancestor of refB
func IsAncestor(refA string, refB string) (bool, error) {
	cmd := exec.Command("git", "merge-base", "--is-ancestor", refA, refB)
	if err := cmd.Run(); err != nil {
		// --is-ancestor returns exit code 1 if the commit is not a part of main.
		if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() == 1 {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func MergeBase(refA string, refB string) (_ string, onMain bool, _ error) {
	mergeBaseOut, err := exec.Command("git", "merge-base", refA, refB).Output()
	if err != nil {
		return "", false, err
	}
	hash := strings.TrimSpace(string(mergeBaseOut))
	isOnMain, err := IsAncestor(hash, mainBranch)
	if err != nil {
		return hash, false, err
	}
	return hash, isOnMain, nil
}

func CurrentBranch() (string, error) {
	out, err := exec.Command("git", "symbolic-ref", "HEAD").Output()
	if err != nil {
		return "", err
	}
	nameMatch := refHeadRegexp.FindStringSubmatch(strings.TrimSpace(string(out)))
	if nameMatch == nil {
		return "", fmt.Errorf("unexpected ref name: \"%s\"", string(out))
	}
	return nameMatch[1], nil
}
