package pipelineWorker

import (
	"fmt"
	"log"
	"time"
)

type JobChannel chan Job
type JobQueue chan chan Job

// Worker is a a single processor. Typically its possible to
// start multiple workers for better throughput
type Worker struct {
	ID      int           // id of the worker
	Name    string        //name of the worker
	JobChan JobChannel    // a channel to receive single unit of work
	Queue   JobQueue      // shared between all workers.
	Quit    chan struct{} // a channel to quit working
}

func New(ID int, Name string, JobChan JobChannel, Queue JobQueue, Quit chan struct{}) *Worker {
	return &Worker{
		ID:      ID,
		Name:    Name,
		JobChan: JobChan,
		Queue:   Queue,
		Quit:    Quit,
	}
}

func (wr *Worker) Start() {
	//c := &http.Client{Timeout: time.Millisecond * 15000}
	go func() {
		for {
			done := make(chan struct{})
			Machines := make(chan Machine, 2)
			defer close(done)
			// when available, put the JobChan again on the JobPool
			// and wait to receive a job
			wr.Queue <- wr.JobChan
			select {
			case job := <-wr.JobChan:
				for machine := range Machines {
					wr.Pipeline(done, machine, job)
				}
				// when a job is received, process
				//callApi(job.ID, wr.ID, c)
				//job.Dojob(wr.ID, job)
				//job.JobID()

			case <-wr.Quit:
				// a signal on this channel means someone triggered
				// a shutdown for this worker
				close(wr.JobChan)
				return
			}
		}
	}()
}

// stop closes the Quit channel on the worker.
func (wr *Worker) Stop() {
	close(wr.Quit)
}

func (wr *Worker) Pipeline(done <-chan struct{}, machine Machine, job Job) <-chan Job {
	start := time.Now()
	wr.Name = machine.name()
	prefix := fmt.Sprintf("Worker[%d::%s]", wr.ID, wr.Name)
	postfix := fmt.Sprintf("Job[%d::%s]", job.JobID(), job.JobName())
	fmt.Println(prefix, "start to do", postfix, "!")
	c := make(chan Job)
	time.Sleep(time.Millisecond * 1)

	c <- SemiFinishedProduct{(job.JobID()), "grindBeanSemiFinishedProduct", time.Now(), time.Now()}

	end := time.Now()
	log.Print(end.Sub(start).Seconds())
	return c
}
