package main

import (
	"fmt"
	"sync"
)

func run() error {
	recentHeads, err := getRecentHeads()
	if err != nil {
		return fmt.Errorf("error getting recent heads: %w", err)
	}
	res := make(chan []string, 10)
	go func() {
		wg := sync.WaitGroup{}
		for _, ref := range recentHeads[:50] {
			wg.Add(1)
			go func() {
				defer wg.Done()
				logs, err := getLogs(ref)
				fmt.Println(ref)
				if err != nil {
					//return fmt.Errorf("error getting logs for ref %s: %w", ref, err)
					fmt.Printf("error getting logs for ref %s: %s", ref, err)
				}
				res <- logs
			}()
		}
		wg.Wait()
		close(res)
	}()
	for _ = range res {
		//fmt.Printf("%v\n", el[:3])
		fmt.Println("done")
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}
