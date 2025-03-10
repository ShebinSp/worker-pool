package work

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

type Executor interface {
	Execute() error
	OnError(error)
}

type Pool struct {
	numWorkers    int
	tasks         chan Executor
	start         sync.Once
	stop          sync.Once
	taskCompleted chan bool
	quit          chan struct{}
}

func NewPool(numWorkers int, taskChannelSize int) (*Pool, error) {
	if numWorkers <= 0 {
		return nil, errors.New("num workers connot be less, or equal to zero")
	}

	if taskChannelSize < 0 {
		return nil, errors.New("channel size cannot be a negative value")
	}

	return &Pool{
		numWorkers:    numWorkers,
		tasks:         make(chan Executor, taskChannelSize),
		start:         sync.Once{},
		stop:          sync.Once{},
		taskCompleted: make(chan bool),
		quit:          make(chan struct{}),
	}, nil
}

func (p *Pool) Start(ctx context.Context) {
	p.start.Do(func() {
		p.startWorker(ctx)
	})
}

func (p *Pool) Stop() {
	p.stop.Do(func() {
		close(p.quit)
		close(p.tasks)
	})
}

func (p *Pool) AddTask(t Executor) {
	select {
	case p.tasks <- t:

	case <-p.quit:
	}
}

// func (p *Pool) AddTaskNonBlocking(t Executor) {
// 	go func() {
// 		p.tasks <- t
// 	}()
// }

func (p *Pool) TaskCompleted() <-chan bool {
	return p.taskCompleted
}

func (p *Pool) startWorker(ctx context.Context) {

	for i := 1; i <= p.numWorkers; i++ {

		go func(workerNum int) {
			fmt.Printf("worker %d started\n", workerNum)
			for {
				select {
				case <-ctx.Done():
					return

				case <-p.quit:
					return

				case task, ok := <-p.tasks:
					if !ok {
						return
					}
					if err := task.Execute(); err != nil {
						task.OnError(err)
					}

					go func() {
						p.taskCompleted <- true
					}()

					fmt.Printf("worker %d finished a task\n", workerNum)
				}
			}
		}(i)

	}
}
