package utils

import (
	"sync"
)

type ImgWorker struct {
}

type Worker interface {
	Do()
}

type Pool[W Worker] struct {
	workers        chan *W
	workerCount    int
	maxConcurrency int
	// totalRuns      int
	initFunc       func() *W
	Wg             sync.WaitGroup
}

func NewPool[W Worker](maxConcurrency int, initF func() *W) Pool[W] {
	workers := make(chan *W, maxConcurrency)
	// lazy worker availability
	for range maxConcurrency {
		workers <- nil
	}

	return Pool[W]{
		workers:        workers,
		workerCount:    0,
		// totalRuns:      totalRuns,
		maxConcurrency: maxConcurrency,
		initFunc:       initF,
	}
}

// func (self *Pool[W]) Run() {
// 	for range self.maxConcurrency {
// 		w := self.get()
// 		go func(w *W) {
// 			fmt.Println("Running routine")
// 			// derefencing here so I can call the Do() method.
// 			// Is there a better way? What is the impact of this?
// 			// Would this deref undermine the suppossed gain of using pointers to
// 			// send workers through the channel?
// 			(*w).Do()
// 			self.wg.Done()
// 			// Putting the worker back in the pool
// 			self.put(w)
// 		}(w)
// 	}
// 	fmt.Println("num workers: ", self.workerCount)
// 	self.wg.Add(self.totalRuns)
// 	// waiting in a separate goroutine so we don't block here
// 	self.wg.Wait()
// 	// go func() {
// 	// }()
// }

func (self *Pool[W]) Get() *W {
	worker := <-self.workers
	if worker == nil {
		worker = self.initFunc()
		self.workerCount++
	}

	return worker
}

func (self *Pool[W]) Put(worker *W) {
	self.workers <- worker
}
