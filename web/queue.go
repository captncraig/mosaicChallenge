package main

import (
	"fmt"
	"log"
	"runtime"
	"sync"
	"time"
)

type job struct {
	id                string
	mainImgUrl        string
	builtInCollection string
	imgurGallery      string
	status, subStatus string
	version           int

	l sync.Mutex
}

func (j *job) UpdateQueuePosition(pos int) {
	j.UpdateStatus("Waiting for an available worker", fmt.Sprintf("position %d in queue", pos))
}

func (j *job) UpdateStatus(status, substatus string) {
	j.l.Lock()
	j.status = status
	j.subStatus = substatus
	j.version++
	j.l.Unlock()
}

// All jobs currently in the pipeline
var allJobs map[string]*job

//channel to submit new jobs on
var jobSubmission = make(chan *job)

func init() {
	go runWorkQueue()
}

func runWorkQueue() {
	//channel to give jobs to workers
	var work = make(chan *job)

	log.Printf("Starting %d worker routines.\n", runtime.NumCPU())
	for i := 0; i < runtime.NumCPU(); i++ {
		go worker(work)
	}

	queue := []*job{}
	enqueue := func(j *job) {
		queue = append(queue, j)
		j.UpdateQueuePosition(len(queue))
	}
	for {
		// if anything in the queue we send and receive concurrently
		// if queue is empty we only receive
		if len(queue) == 0 {
			select {
			case j := <-jobSubmission:
				enqueue(j)
			}
		} else {
			select {
			case j := <-jobSubmission:
				enqueue(j)
			case work <- queue[0]:
				queue = queue[1:]
				for i, j := range queue {
					j.UpdateQueuePosition(i + 1)
				}
			}
		}
	}
}

func worker(q <-chan *job) {
	for {
		j := <-q
		fmt.Println(j.id)
		time.Sleep(3 * time.Second)
	}
}
