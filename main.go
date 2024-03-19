package main

import (
	"fmt"
	"sync"

	"github.com/themichaellai/stacked-branches/git"
)

func getRecentHeads() error {
	recentHeads, errs := git.GetRecentHeads()
	if errs != nil {
		return fmt.Errorf("error getting recent heads: %w", errs)
	}
	res := make(chan []git.Commit, 10)
	errChan := make(chan error)
	go func() {
		wg := sync.WaitGroup{}
		for _, ref := range recentHeads[:50] {
			wg.Add(1)
			go func() {
				defer wg.Done()
				logs, err := git.GetLogs(ref)
				if err != nil {
					errChan <- fmt.Errorf("error getting logs for ref %s: %w", ref, err)
				}
				res <- logs
			}()
		}
		wg.Wait()
		close(res)
	}()
	for {
		select {
		case el, more := <-res:
			if !more {
				fmt.Println("no more")
				return nil
			}
			fmt.Println(el[0])
		case err := <-errChan:
			return err
		}
	}
}

func run() error {
	err := getRecentHeads()
	return err
}

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}
