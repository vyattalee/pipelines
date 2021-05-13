package dispatcher1

import (
	"coffeeshop/worker1"
	"log"
)

// New returns a new dispatcher. A Dispatcher communicates between the client
// and the worker1. Its main job is to receive a job and share it on the WorkPool
// WorkPool is the link between the dispatcher and all the workers as
// the WorkPool of the dispatcher is common JobPool for all the workers
func New(num int) *disp {
	return &disp{
		Workers:      make([]*worker1.Worker, num),
		PipelineChan: make(worker1.PipelineChannel),
		Queue:        make(worker1.PipelineQueue),
	}
}

// disp is the link between the client and the workers
type disp struct {
	Workers      []*worker1.Worker // this is the list of workers that dispatcher tracks
	PipelineChan worker1.PipelineChannel
	Queue        worker1.PipelineQueue // this is the shared JobPool between the workers
}

// Start creates pool of num count of workers.
func (d *disp) Start() *disp {
	l := len(d.Workers)
	for i := 1; i <= l; i++ {
		//wg.Add(1)	//根据pipeline个数设定
		wrk := worker1.New(i, make(worker1.PipelineChannel), make(worker1.JobChannel), d.Queue, make(chan struct{}))
		wrk.Start()
		d.Workers = append(d.Workers, wrk)
	}
	go d.process()

	pipeline1, pipeline2 := d.constructPipeline()

	d.submitPipeline(*pipeline1)
	d.submitPipeline(*pipeline2)

	select {
	case <-pipeline1.PipelineDone:
		log.Println("pipeline1-", pipeline1.Name, " done")
	}
	select {
	case <-pipeline2.PipelineDone:
		log.Println("pipeline2-", pipeline2.Name, " done")
	}

	return d
}

// process listens to a job submitted on WorkChan and
// relays it to the WorkPool. The WorkPool is shared between
// the workers.
func (d *disp) process() {
	for {
		select {
		case pipeline := <-d.PipelineChan: // listen to any submitted job on the WorkChan
			// wait for a worker1 to submit JobChan to JobQueue
			// note that this JobQueue is shared among all workers.
			// Whenever there is an available JobChan on JobQueue pull it
			pipelineChan := <-d.Queue

			// Once a jobChan is available, send the submitted Job on this JobChan
			pipelineChan <- pipeline
		}
	}
}

func (d *disp) constructPipeline() (*worker1.Pipeline, *worker1.Pipeline) {
	pipeline1 := worker1.NewPipeline(1, "grindBean_espressoCoffee_pipeline")
	pipeline1.Machines <- &worker1.GrindBeanMachine{}
	pipeline1.Machines <- &worker1.EspressoCoffeeMachine{}
	close(pipeline1.Machines) //非常重要，不用了的channel务必关闭掉，否则就会有deadlock，继续等待channel接收数据

	pipeline2 := worker1.NewPipeline(2, "steamMilk_pipeline")
	pipeline2.Machines <- &worker1.SteamMilkMachine{}
	close(pipeline2.Machines) //非常重要，不用了的channel务必关闭掉，否则就会有deadlock，继续等待channel接收数据
	return pipeline1, pipeline2
}

//func (d *disp) SubmitJob(job worker1.Job) {
//	d.JobChan <- job
//}

func (d *disp) submitPipeline(pipeline worker1.Pipeline) {
	d.PipelineChan <- pipeline
}
