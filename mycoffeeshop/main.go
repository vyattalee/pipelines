package main

import (
	"coffeeshop/dispatcher"
	"coffeeshop/dispatcher1"
	"coffeeshop/dispatcher2"
	pipeline "coffeeshop/pipeline_model"
	"context"
	"io"
	"io/ioutil"
	"net/http"
	//"coffeeshop/pipelineDispatcher"
	//"coffeeshop/pipelineWorker"
	"coffeeshop/worker"
	"coffeeshop/worker2"
	"fmt"
	"log"
	"math/rand"
	"time"
)

func main() {
	//dispatch_worker()

	//dispatch_worker1()

	//dispatch_worker2()

	//test4PolymorphicInheritanceBYInterface()
	//pipelineDispatch_worker()

	//pipeline_stage_step_model()
	//
	//simple_pipeline_model()

	coffeeshop_pipeline_model(false)

	//coffeeshop_pipeline_model(true)
}
func test4PolymorphicInheritanceBYInterface() {
	var m Mammal
	var mammalchan chan Mammal

	m = &Dog{}
	m.Say()
	m = &Cat{}
	m.Say()
	m = &Human{}
	m.Say()

	mammalchan = make(chan Mammal, 5)
	mammalchan <- Human{}
	mammalchan <- Dog{}
	mammalchan <- Cat{}
	close(mammalchan) //remember channel should be closed after useing up

	for m := range mammalchan {
		//log.Println(m.Say())
		m.Say()
	}
}
func dispatch_worker() {
	start := time.Now()
	dd := dispatcher.New(10).Start()

	terms := map[int]string{
		1:  "grindBean",
		2:  "espressoCoffee",
		3:  "steamMilk",
		4:  "grindBean",
		5:  "espressoCoffee",
		6:  "steamMilk",
		7:  "grindBean",
		8:  "espressoCoffee",
		9:  "steamMilk",
		10: "grindBean",
		11: "espressoCoffee",
		12: "steamMilk",
		13: "grindBean",
		14: "espressoCoffee",
		15: "steamMilk",
		17: "grindBean",
		18: "espressoCoffee",
		19: "steamMilk",
		16: "coffeeLatte"}

	for id, name := range terms {
		dd.Submit(worker.Job{
			ID:   id,
			Name: fmt.Sprintf("Job-::%s", name),
			Dojob: func(id int, job worker.Job) {
				start := time.Now()
				prefix := fmt.Sprintf("##Worker[%d]-Job[%d::%s]", id, job.ID, job.Name)
				log.Print(prefix, "start to do job!")
				time.Sleep(time.Second * time.Duration(rand.Intn(10)))
				end := time.Now()
				log.Print(prefix, "finish job total time: ", end.Sub(start).Seconds())
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		})
	}

	//dd.Submit(worker.grindBeanJob{
	//	ID:        id,
	//	name:      fmt.Sprintf("Job-::%s", name),
	//	CreatedAt: time.Now(),
	//	UpdatedAt: time.Now(),
	//})

	end := time.Now()
	log.Print(end.Sub(start).Seconds())
}

func dispatch_worker1() {
	start := time.Now()

	//var wg sync.WaitGroup

	/*dd :=*/
	dispatcher1.New(6).Start()

	//grindBean_espressoCoffee_machine := (worker1.grindBeanMachine)

	//terms := map[int]string{
	//	1: "steamMilk_pipeline",
	//	2: "grindBean_espressoCoffee_pipeline",
	//2:  "grindBean_espressoCoffee_pipeline",
	//3:  "steamMilk_pipeline",
	//4:  "grindBean_espressoCoffee_pipeline",
	//5:  "grindBean_espressoCoffee_pipeline",
	//6:  "steamMilk_pipeline",
	//7:  "grindBean_espressoCoffee_pipeline",
	//8:  "grindBean_espressoCoffee_pipeline",
	//9:  "steamMilk_pipeline",
	//10: "grindBean_espressoCoffee_pipeline",
	//11: "grindBean_espressoCoffee_pipeline",
	//12: "steamMilk_pipeline",
	//13: "grindBean_espressoCoffee_pipeline",
	//14: "grindBean_espressoCoffee_pipeline",
	//15: "steamMilk_pipeline",
	//16: "steamMilk_pipeline",
	//17: "grindBean_espressoCoffee_pipeline",
	//18: "grindBean_espressoCoffee_pipeline",
	//}

	//for id, name := range terms {
	//
	//	pipeline := worker1.NewPipeline(id, name)
	//	//pipeline.Machines[0] = worker1.Machine(&worker1.GrindBeanMachine{})
	//	//pipeline.Machines[1] = &worker1.EspressoCoffeeMachine{}
	//	pipeline.Machines <- &worker1.GrindBeanMachine{}
	//	pipeline.Machines <- &worker1.EspressoCoffeeMachine{}
	//
	//	dd.submitPipeline(*pipeline)
	//}

	//pipeline3 := worker1.NewPipeline(3, "3_pipeline")
	//pipeline3.Machines <- &worker1.SteamMilkMachine{}
	//pipeline4 := worker1.NewPipeline(4, "4_pipeline")
	//pipeline4.Machines <- &worker1.SteamMilkMachine{}
	//pipeline5 := worker1.NewPipeline(5, "5_pipeline")
	//pipeline5.Machines <- &worker1.SteamMilkMachine{}
	//pipeline6 := worker1.NewPipeline(6, "6_pipeline")
	//pipeline6.Machines <- &worker1.SteamMilkMachine{}
	//pipeline7 := worker1.NewPipeline(7, "7_pipeline")
	//pipeline7.Machines <- &worker1.SteamMilkMachine{}
	//pipeline8 := worker1.NewPipeline(8, "8_pipeline")
	//pipeline8.Machines <- &worker1.SteamMilkMachine{}

	//dd.submitPipeline(*pipeline3)
	//dd.submitPipeline(*pipeline4)
	//dd.submitPipeline(*pipeline5)
	//dd.submitPipeline(*pipeline6)
	//dd.submitPipeline(*pipeline7)
	//dd.submitPipeline(*pipeline8)

	//l := len(dd.Workers)
	//for i := 0; i < l; i++ {
	//	<- dd.Workers[i].Quit
	//}

	end := time.Now()
	log.Print(end.Sub(start).Seconds())
	//wg.Wait()
	//fmt.Scanln()
}

func dispatch_worker2() {
	start := time.Now()

	//var wg sync.WaitGroup

	/**/
	dd := dispatcher2.New(6).Start()

	pipeline1 := worker2.NewPipeline(1, "grindBean_espressoCoffee_pipeline")
	pipeline1.Machines <- &worker2.GrindBeanMachine{}
	pipeline1.Machines <- &worker2.EspressoCoffeeMachine{}
	close(pipeline1.Machines) //非常重要，不用了的channel务必关闭掉，否则就会有deadlock，继续等待channel接收数据

	//pipeline2 := worker2.NewPipeline(2, "steamMilk_pipeline")
	//pipeline2.Machines <- &worker2.SteamMilkMachine{}
	//close(pipeline2.Machines) //非常重要，不用了的channel务必关闭掉，否则就会有deadlock，继续等待channel接收数据
	//
	//pipeline3 := worker2.NewPipeline(3, "haha_pipeline")
	//pipeline3.Machines <- &worker2.HahaMachine{}
	//close(pipeline3.Machines) //非常重要，不用了的channel务必关闭掉，否则就会有deadlock，继续等待channel接收数据

	dd.SubmitPipeline(*pipeline1)
	//dd.SubmitPipeline(*pipeline2)
	//dd.SubmitPipeline(*pipeline3)

	dd.SubmitJob(worker2.Job{
		ID:        1000,
		Name:      fmt.Sprintf("Job-::%s", "Ginger's order"),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})

	dd.SubmitJob(worker2.Job{
		ID:        1001,
		Name:      fmt.Sprintf("Job-::%s", "Johan's order"),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})

	dd.SubmitJob(worker2.Job{
		ID:        1002,
		Name:      fmt.Sprintf("Job-::%s", "Vyatta's order"),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})

	dd.SubmitJob(worker2.Job{
		ID:        1003,
		Name:      fmt.Sprintf("Job-::%s", "Noha's order"),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})

	//dd.SubmitJob(worker2.SemiFinishedProduct{
	//	ProductId:          1003,
	//	ProductDescription: fmt.Sprintf("Job-::%s", "semiFinishedProduct"),
	//})

	end := time.Now()
	log.Print(end.Sub(start).Seconds())
	//wg.Wait()
	//fmt.Scanln()
}

func pipelineDispatch_worker() {
	start := time.Now()
	//dd := pipelineDispatcher.New(10).Start()

	//Job 主要是订单
	//terms := map[int]string{
	//	1000: "grindBean",
	//	2:    "espressoCoffee",
	//	3:    "steamMilk",
	//	4:    "grindBean",
	//	5:    "espressoCoffee",
	//	6:    "steamMilk",
	//	7:    "grindBean",
	//	8:    "espressoCoffee",
	//	9:    "steamMilk",
	//	10:   "grindBean",
	//	11:   "espressoCoffee",
	//	12:   "steamMilk",
	//	13:   "grindBean",
	//	14:   "espressoCoffee",
	//	15:   "steamMilk",
	//	17:   "grindBean",
	//	18:   "espressoCoffee",
	//	19:   "steamMilk",
	//	16:   "coffeeLatte"}

	//for id, name := range terms {
	//	//dd.Submit(pipelineWorker.Job{
	//	//	ID:   id,
	//	//	NAME: fmt.Sprintf("Job-::%s", name),
	//	//	//Dojob: func(id int, job pipelineWorker.Job) {
	//	//	//	start := time.Now()
	//	//	//	prefix := fmt.Sprintf("##Worker[%d]-Job[%d::%s]", id, job.ID, job.NAME)
	//	//	//	log.Print(prefix, "start to do job!")
	//	//	//	time.Sleep(time.Second * time.Duration(rand.Intn(10)))
	//	//	//	end := time.Now()
	//	//	//	log.Print(prefix, "finish job total time: ", end.Sub(start).Seconds())
	//	//	//},
	//	//	CreatedAt: time.Now(),
	//	//	UpdatedAt: time.Now(),
	//	//})
	//}

	end := time.Now()
	log.Print(end.Sub(start).Seconds())
}

type CoffeeStep struct {
	pipeline.StepContext
	NAME        string
	ProcessTime int64
	fail        bool
	ctx         context.Context
	cancel      context.CancelFunc
}

func newStep(Name string, ProcessTime int64, fail bool) *CoffeeStep {
	ctx, cancel := context.WithCancel(context.Background())
	d := &CoffeeStep{NAME: Name, ProcessTime: ProcessTime, fail: fail}
	d.ctx = ctx
	d.cancel = cancel
	return d
}

func (d *CoffeeStep) Name() string {
	return d.NAME
}

func (d *CoffeeStep) Exec(request *pipeline.Request) *pipeline.Result {

	d.Status(fmt.Sprintf("%+v", request))

	d.Status(fmt.Sprintf("Started %s", d.NAME))

	time.Sleep(time.Millisecond * time.Duration(d.ProcessTime))

	if d.fail {
		return &pipeline.Result{Error: fmt.Errorf("step failed %s", d.Name)}
	}

	d.Status(fmt.Sprintf("Successfully %s", d.NAME))

	return &pipeline.Result{
		Error:  nil,
		Data:   request.Data, /*struct{ SemiFinishedProductID int64 }{SemiFinishedProductID: }*/
		KeyVal: map[string]interface{}{"SemiFinishedProduct": request.Data},
	}
}

/*
 &{Data:{Order:1000} KeyVal:map[Customer Order:1000]}
request.KeyVal    &{Data:{Order:1000} KeyVal:map[SemiFinishedProductID:map[SemiFinishedProductID:map[Customer Order:1000]]]}
request.Data	  &{Data:{Order:1000} KeyVal:map[Customer Order:1000]}

*/

func (d *CoffeeStep) Cancel() error {
	d.Status(fmt.Sprintf("Cancel %s", d.Name))
	d.cancel()
	return nil
}

type GrindBeanStep struct {
	pipeline.StepContext
	Name string
	fail bool

	ctx    context.Context
	cancel context.CancelFunc
}

func coffeeshop_pipeline_model(isConcurrent bool) {

	start := time.Now()

	for id := int64(0); id < 10; id++ {

		workflow := pipeline.NewProgress("coffee shop", 1000, time.Millisecond*500)

		//stage
		stage := pipeline.NewStage("CoffeeShopStage1", isConcurrent, false)

		//steps
		grindBeanStep := newStep("GrindBeanStep", 100, false)
		espressoCoffeeStep := newStep("EspressoCoffeeStep", 200, false)
		steamMilkStep := newStep("SteamMilkStep", 300, false)

		//stage with sequential steps
		stage.AddStep(grindBeanStep, espressoCoffeeStep, steamMilkStep)

		// add all stages
		workflow.AddStage(stage)

		//channel solution: introduce the Order channel into pipeline to dispatch the order
		//workflow.Submit(pipeline.CustomerOrder{ID: 1000, NAME: "Lattee", CreatedAt: time.Now(), UpdatedAt: time.Now()})
		//workflow.Submit(pipeline.CustomerOrder{ID: 1001, NAME: "Lattee1", CreatedAt: time.Now(), UpdatedAt: time.Now()})
		//workflow.Submit(pipeline.CustomerOrder{ID: 1002, NAME: "Lattee2", CreatedAt: time.Now(), UpdatedAt: time.Now()})

		// start a routine to read out and progress
		go readPipeline(workflow)

		//var request *pipeline.Request
		//for job := range workflow.WorkChan{
		//request batch in for loop solution

		//	request = &pipeline.Request{Data: struct{ Order int64 }{Order: job.JobID()}, KeyVal: map[string]interface{}{job.JobName(): job.JobID()}}

		// execute pipeline
		result := workflow.RunWithID(id)
		//result := workflow.Run()
		//result := workflow.RunWithReq(request)

		if result.Error != nil {
			fmt.Println(result.Error)
		}

		//}
		fmt.Println("timeTaken:", workflow.GetDuration())
	}

	//if call workflow.RunWithID, it should call ClearBufferMap;
	// but other functions(workflow.Run only one request) don't need to do that
	//workflow.ClearBufferMap()

	//close(workflow.WorkChan)
	//close(workflow.Queue)

	// one would persist the time taken duration to use as progress scale for the next workflow build

	end := time.Now()

	log.Print(end.Sub(start).Seconds(), " seconds")

}

func pipeline_stage_step_model() {
	workflow := pipeline.NewProgress("getfiles", 10000, time.Second*7)
	//stages
	stage := pipeline.NewStage("stage", false, false)
	// in this stage, steps will be executed concurrently
	concurrentStage := pipeline.NewStage("con_stage", true, false)
	// another concurrent stage
	concurrentErrStage := pipeline.NewStage("con_err_stage", true, false)

	//steps
	fileStep1mb := newDownloadStep("1mbfile", 1e6, false)
	fileStep1mbFail := newDownloadStep("1mbfileFail", 1e6, true)
	fileStep5mb := newDownloadStep("5mbfile", 5e6, false)
	fileStep10mb := newDownloadStep("10mbfile", 10e6, false)

	//stage with sequential steps
	stage.AddStep(fileStep1mb, fileStep5mb, fileStep10mb)

	//stage with concurrent steps
	concurrentStage.AddStep(fileStep1mb, fileStep5mb, fileStep10mb)

	//stage with concurrent steps one of which fails early, prompting a cancellation
	//of the other running steps.
	concurrentErrStage.AddStep(fileStep1mbFail, fileStep5mb, fileStep10mb)

	// add all stages
	workflow.AddStage(stage, concurrentStage, concurrentErrStage)

	// start a routine to read out and progress
	go readPipeline(workflow)

	// execute pipeline
	result := workflow.Run()
	if result.Error != nil {
		fmt.Println(result.Error)
	}

	// one would persist the time taken duration to use as progress scale for the next workflow build
	fmt.Println("timeTaken:", workflow.GetDuration())
}

type downloadStep struct {
	pipeline.StepContext
	fileName string
	bytes    int64
	fail     bool
	ctx      context.Context
	cancel   context.CancelFunc
}

func newDownloadStep(fileName string, bytes int64, fail bool) *downloadStep {
	ctx, cancel := context.WithCancel(context.Background())
	d := &downloadStep{fileName: fileName, bytes: bytes, fail: fail}
	d.ctx = ctx
	d.cancel = cancel
	return d
}
func (d *downloadStep) Name() string {
	return d.fileName
}
func (d *downloadStep) Exec(request *pipeline.Request) *pipeline.Result {

	d.Status(fmt.Sprintf("%+v", request))

	d.Status(fmt.Sprintf("Started downloading file %s", d.fileName))

	client := &http.Client{}

	req, err := http.NewRequest("GET", fmt.Sprintf("http://httpbin.org/ProcessTime/%d", d.bytes), nil)
	if err != nil {
		return &pipeline.Result{Error: err}
	}

	req = req.WithContext(d.ctx)

	resp, err := client.Do(req)
	if err != nil {
		return &pipeline.Result{Error: err}
	}

	defer resp.Body.Close()

	n, err := io.Copy(ioutil.Discard, resp.Body)
	if err != nil {
		return &pipeline.Result{Error: err}
	}

	if d.fail {
		return &pipeline.Result{Error: fmt.Errorf("File download failed %s", d.fileName)}
	}

	d.Status(fmt.Sprintf("Successfully downloaded file %s", d.fileName))

	return &pipeline.Result{
		Error:  nil,
		Data:   struct{ bytesDownloaded int64 }{bytesDownloaded: n},
		KeyVal: map[string]interface{}{"ProcessTime": n},
	}
}

func (d *downloadStep) Cancel() error {
	d.Status(fmt.Sprintf("Cancel downloading file %s", d.fileName))
	d.cancel()
	return nil
}

//var m sync.Mutex
func readPipeline(pipe *pipeline.Pipeline) {
	//m.Lock()
	//defer m.Unlock()
	out, err := pipe.Out()
	if err != nil {
		return
	}

	progress, err := pipe.GetProgressPercent()
	if err != nil {
		return
	}

	for {
		select {
		case line := <-out:
			fmt.Println(line)
		case p := <-progress:
			fmt.Println("percent done: ", p)
		}
	}
}

/////////////////////////following simple example////////////////////////

func simple_pipeline_model() {

	workpipe := pipeline.NewProgress("myProgressworkpipe", 1000, time.Second*3)
	stage := pipeline.NewStage("mypworkstage", false, false)
	step1 := &work{id: 1}
	step2 := &work{id: 2}

	stage.AddStep(step1)
	stage.AddStep(step2)

	workpipe.AddStage(stage)

	go readPipeline(workpipe)

	result := workpipe.Run()
	if result.Error != nil {
		fmt.Println(result.Error)
	}

	fmt.Println("timeTaken:", workpipe.GetDuration())
}

type work struct {
	pipeline.StepContext
	id int
}

func (w *work) Name() string {
	return w.Name()
}
func (w *work) Exec(request *pipeline.Request) *pipeline.Result {

	w.Status(fmt.Sprintf("%+v", request))
	duration := time.Duration(1000 * w.id)
	time.Sleep(time.Millisecond * duration)
	msg := fmt.Sprintf("work %d", w.id)
	return &pipeline.Result{
		Error:  nil,
		Data:   map[string]string{"msg": msg},
		KeyVal: map[string]interface{}{"msg": msg},
	}
}

func (w *work) Cancel() error {
	w.Status("cancel step")
	return nil
}
