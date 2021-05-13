//worker1 binding the unique machine, such as grindBean/espressoCoffee/steamMilk machine
//

package worker2

import (
	"log"
	"time"
)

// Job represents a single entity that should be processed.
// For example a struct that should be saved to database

type JobQueue chan chan JobInterface

//type JobChannel chan Job interface
type JobChannel chan JobInterface

type PipelineQueue chan chan Pipeline

type PipelineChannel chan Pipeline

// Worker is a a single processor. Typically its possible to
// start multiple workers for better throughput
type Worker struct {
	ID   int    // id of the worker
	Name string //name of the worker
	//PipelineChan JobChannel    //a channel to machine list, a worker1 can deal with several machine in future
	PipelineQueue PipelineQueue // shared between all workers
	PipelineChan  PipelineChannel
	JobQueue      JobQueue      // shared between all workers.
	JobChan       JobChannel    // a channel to
	Quit          chan struct{} // a channel to quit working
}

func New(ID int, Name string, PipelineChan PipelineChannel, PipelineQueue PipelineQueue, JobChan JobChannel, JobQueue JobQueue, Quit chan struct{}) *Worker {
	return &Worker{
		ID:            ID,
		Name:          Name,
		PipelineChan:  PipelineChan,
		PipelineQueue: PipelineQueue,
		JobChan:       JobChan,
		JobQueue:      JobQueue,
		Quit:          Quit,
	}
}

func (wr *Worker) Start() {
	//c := &http.Client{Timeout: time.Millisecond * 15000}
	go func() {
		//defer wg.Done()
		for {

			// when available, put the JobChan again on the JobPool
			// and wait to receive a job
			wr.PipelineQueue <- wr.PipelineChan

			select {
			case pipeline := <-wr.PipelineChan:
				wr.JobQueue <- wr.JobChan
				select {
				case job := <-wr.JobChan:
					//log.Println("Worker[", wr.ID, "::", wr.NAME, "] do job[", job.JobID(), "::", job.JobName(), "]")
					for machine := range pipeline.Machines {
						job = machine.dojob(*wr, job)
						log.Println("semiFinishedProduct[", job.JobID(), "::", job.JobName(), "]")
						//semiFinishedProduct := machine.dojob(*wr, Job{int64(pipeline.ID<<6 + wr.ID), pipeline.NAME + machine.name(), time.Now(), time.Now()})
						//log.Println("semiFinishedProduct[", semiFinishedProduct.ProductId, "::", semiFinishedProduct.ProductDescription, "]")

					}
					//pipeline.PipelineDone <- struct{}{}
					//close(pipeline.PipelineDone)

				case <-wr.Quit:
					// a signal on this channel means someone triggered
					// a shutdown for this worker
					close(wr.JobChan)
					return
				}
				//wr.PipelineChan <- pipeline

			case timeout := <-time.After(time.Second * 1):
				log.Println("timeout", timeout)

			}

		} //end of for

	}()
}

// stop closes the Quit channel on the worker.
func (wr *Worker) Stop() {
	close(wr.Quit)
	close(wr.JobChan)
}

//func (wr *Worker) SubmitJob(job JobInterface) {
//	wr.JobChan <- job
//}

//func callPipeline(pipeline dispatcher2.Pipeline) {
//	log.Println("pipeline:", pipeline.NAME, "@", pipeline.ID, "  createAt:", pipeline.CreatedAt, "  updateAt:", pipeline.UpdatedAt)
//}
