package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/ShebinSp/worker-pool/work"
)

func main() {
	wp, err := work.NewPool(5, 5)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wp.Start(ctx)


	for range 20 {
		task := work.NewTask(func() error {
			const urlString = "https://google.com"
			resp, err := http.Get(urlString)
			if err != nil {
				return err
			}

			fmt.Printf("%s returned status code %d\n", urlString, resp.StatusCode)

			return nil
		}, func(err error) {
			log.Println(err)
		})

		// wp.AddTaskNonBlocking(task)
		wp.AddTask(task)
	}

	counter := 0
	
	for completed := range wp.TaskCompleted() {
		if completed {
			counter++
		}


		if counter == 20 {
			wp.Stop()
			fmt.Printf("\nTotal number of tasks completed: %d\n", counter)
			return
		}
	}
	
}

