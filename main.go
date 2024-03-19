package main

import (
	"fmt"
	"sync"

	"github.com/themichaellai/stacked-branches/git"
)

const concurrency = 10

func getMergeBases(targetRef string, refs []string) (refToMergeBase map[string]string, _ error) {
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
				mergeBase, onMain, err := git.MergeBase(targetRef, ref)
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

type Node struct {
	Hash string
	// Children are commits that are newer
	Children []string
	// Parent is the commit that is older
	Parent string
}

//func getCandidateStack(logs map[string]([]git.Commit), targetRef string) ([]string, error) {
//	// Build commit tree
//	//hashTree := map[string]([]string){}
//	//refToHead := map[string]string{}
//	//for ref, log := range logs {
//	//	fmt.Println(ref)
//	//	refToHead[ref] = log[0].Hash
//	//	for idx, commit := range log[:len(log)-1] {
//	//		hashTree[commit.Hash] = append(hashTree[commit.Hash], log[idx+1].Hash)
//	//	}
//	//}
//	tree := map[string]*Node{}
//	for ref, log := range logs {
//		fmt.Println(ref)
//		for idx, commit := range log[:len(log)-1] {
//			node := tree[commit.Hash]
//			fmt.Printf("curr hash: %s\n", commit.Hash)
//			if node == nil {
//				node = &Node{Hash: commit.Hash}
//				tree[commit.Hash] = node
//			}
//			if node.Parent != "" && node.Parent != log[idx+1].Hash {
//				fmt.Printf("node.Parent: %s, log[idx+1].Hash: %s\n", node.Parent, log[idx+1].Hash)
//				return nil, fmt.Errorf("foo")
//			}
//			node.Parent = log[idx+1].Hash
//			if idx != 0 {
//				node.Children = append(node.Children, log[idx-1].Hash)
//			}
//		}
//	}
//	for _, commit := range logs[targetRef] {
//		fmt.Println(commit)
//	}
//
//	// Find merge base for target ref with all other refs
//	// Do this by commit hash, but in future can use commit message as well
//	// Order refs with the highest number of common commits into stack
//
//	return nil, nil
//}

func getCandidateStack(targetRef string) ([]string, error) {
	recentHeads, err := git.GetRecentHeads()
	if err != nil {
		return nil, fmt.Errorf("error getting recent heads: %w", err)
	}
	mergeBases, err := getMergeBases(targetRef, recentHeads)
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
	currentBranch, err := git.CurrentBranch()
	if err != nil {
		return fmt.Errorf("could not get current branch: %w", err)
	}
	stack, err := getCandidateStack(currentBranch)
	if err != nil {
		return fmt.Errorf("error building stack: %w", err)
	}
	fmt.Printf("stack: %#v\n", stack)
	return err
}

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}
