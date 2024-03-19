package main

import (
	"fmt"
	"os"
	"slices"
	"sync"

	"github.com/themichaellai/stacked-branches/git"
)

const concurrency = 10

func getMergeBases(gitClient git.Client, targetRef string, refs []string) (refToMergeBase map[string]string, _ error) {
	type result struct {
		ref  string
		base string
	}
	workRefs := make(chan string, 10)
	resChan := make(chan result)
	errChan := make(chan error)
	doneChan := make(chan struct{})
	wg := sync.WaitGroup{}
	go func() {
		for _, ref := range refs {
			workRefs <- ref
		}
		close(workRefs)
	}()
	for range concurrency {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for ref := range workRefs {
				mergeBase, onMain, err := gitClient.MergeBase(targetRef, ref)
				if err != nil {
					errChan <- err
					return
				}
				if onMain {
					continue
				}
				resChan <- result{ref, mergeBase}
			}
		}()
	}
	go func() {
		wg.Wait()
		close(doneChan)
	}()
	refToMergeBase = map[string]string{}
	for {
		select {
		case el := <-resChan:
			refToMergeBase[el.ref] = el.base
		case err := <-errChan:
			return nil, err
		case <-doneChan:
			return refToMergeBase, nil
		}
	}
}

func sortRefs(gitClient git.Client, refs []string, targetRef string) error {
	// We neeed to compare the merge base with each ref rather than the ref
	// directly because the ref might have diverged from the target ref.
	refMergeBaseWithTarget := map[string]string{}
	for _, ref := range refs {
		mergeBase, _, err := gitClient.MergeBase(ref, targetRef)
		if err != nil {
			return err
		}
		refMergeBaseWithTarget[ref] = mergeBase
	}
	var err error
	slices.SortFunc(refs, func(a, b string) int {
		if err != nil {
			return 0
		}
		var isAncestor bool
		if a == b {
			return 0
		}
		isAncestor, err = gitClient.IsAncestor(refMergeBaseWithTarget[a], refMergeBaseWithTarget[b])
		if isAncestor {
			return -1
		}
		return 1
	})
	return err
}

// TODO: to get branches that are based off this branch, look for other
// branches that contain commits on the current target ref / have a merge base
// that is not on main
func getCandidateStack(gitClient git.Client, targetRef string) ([]string, error) {
	recentHeads, err := git.GetRecentHeads()
	if err != nil {
		return nil, fmt.Errorf("error getting recent heads: %w", err)
	}
	mergeBases, err := getMergeBases(gitClient, targetRef, recentHeads)
	if err != nil {
		return nil, fmt.Errorf("error getting merge bases: %w", err)
	}
	res := make([]string, 0, len(mergeBases))
	for ref := range mergeBases {
		res = append(res, ref)
	}
	return res, nil
}

func run() error {
	mainBranch := "origin/main"
	if mainEnv, setMain := os.LookupEnv("GIT_MAIN"); setMain {
		mainBranch = mainEnv
	}
	gitClient := git.Client{MainBranch: mainBranch}
	currentBranch, err := gitClient.CurrentBranch()
	if err != nil {
		return fmt.Errorf("could not get current branch: %w", err)
	}
	stackRefs, err := getCandidateStack(gitClient, currentBranch)
	if err != nil {
		return fmt.Errorf("error building stack: %w", err)
	}
	if err := sortRefs(gitClient, stackRefs, currentBranch); err != nil {
		return fmt.Errorf("error sorting: %w", err)
	}
	slices.Reverse(stackRefs)
	for _, stackRef := range stackRefs {
		fmt.Println(stackRef)
	}
	return err
}

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}
