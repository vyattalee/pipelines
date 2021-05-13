package pipeline

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/fatih/color"
)

// DefaultDrainTimeout time to wait for all readers to finish consuming output
const DefaultDrainTimeout = time.Second * 5

// DefaultBuffer channel buffer size of the output buffer
const DefaultBuffer = 1000

type JobChannel chan Job
type JobQueue chan chan Job

// Pipeline is a sequence of stages
type Pipeline struct {
	Name             string   `json:"name"`
	Stages           []*Stage `json:"stages"`
	DrainTimeout     time.Duration
	expectedDuration time.Duration
	duration         time.Duration
	outsubscribed    bool
	outbufferlen     int
	tick             time.Duration
	cancelDrain      context.CancelFunc
	cancelProgress   context.CancelFunc
	WorkChan         JobChannel // client submits job to this channel
	Queue            JobQueue   // this is the shared JobPool between the workers
}

// New returns a new pipeline
// 	name of the pipeline
// 	outBufferLen is the size of the output buffered channel
func New(name string, outBufferLen int) *Pipeline {

	return newPipeline(name, outBufferLen)
}

// NewProgress returns a new pipeline which returns progress updates
// 	name of the pipeline
// 	outBufferLen is the size of the output buffered channel
//
// 	expectedDurationInMs is the expected time for the job to finish in milliseconds
// 	If set, you can get the current time spent from GetDuration()int64 and
// 	listen on the channel returned by GetProgress() <-chan float64 to get current progress
func NewProgress(name string, outBufferLen int, expectedDuration time.Duration) *Pipeline {

	p := newPipeline(name, outBufferLen)
	p.expectedDuration = expectedDuration
	p.tick = time.Millisecond * 250
	return p
}

func newPipeline(name string, outBufferLen int) *Pipeline {
	if outBufferLen < 0 {
		outBufferLen = 1
	}

	if buffersMap == nil {
		buffersMap = &buffers{bufferMap: make(map[string]*buffer)}
	}

	//p := &Pipeline{NAME: spaceMap(name)/*, WorkChan: make(JobChannel), Queue: make(JobQueue)*/}
	p := &Pipeline{Name: spaceMap(name), WorkChan: make(JobChannel, 100), Queue: make(JobQueue)}
	p.outbufferlen = outBufferLen

	if p.DrainTimeout == 0 {
		p.DrainTimeout = DefaultDrainTimeout
	}

	buf := buffer{in: make(chan string, outBufferLen), out: []chan string{}, progress: []chan int64{}}
	buffersMap.set(p.Name, &buf)

	return p
}

// SetDrainTimeout sets DrainTimeout
func (p *Pipeline) SetDrainTimeout(timeout time.Duration) {
	p.DrainTimeout = timeout
}

// AddStage adds a new stage to the pipeline
func (p *Pipeline) AddStage(stage ...*Stage) {
	for i := range stage {
		for j := range stage[i].Steps {
			ctx := &stepContextVal{
				name:        p.Name + "." + stage[i].Name + "." + reflect.TypeOf(stage[i].Steps[j]).String(),
				pipelineKey: p.Name,
				concurrent:  stage[i].Concurrent,
				index:       j,
			}

			stage[i].Steps[j].setCtx(ctx)
		}
		stage[i].pipelineKey = p.Name
	}

	p.Stages = append(p.Stages, stage...)
}

//if call workflow.RunWithID, it should call ClearBufferMap;
// but other functions(workflow.Run only one request) don't need to do that
// it is no use any more, because the pipeline are recreated from scratch NewProgress/NewStage/NewStep/AddStep
func (p *Pipeline) ClearBufferMap() {
	buffersMap.remove(p.Name) //after RunWithID in loop, it should ClearBufferMap
}

// Run the pipeline. The stages are executed in sequence while steps may be concurrent or sequential.
func (p *Pipeline) RunWithID(ID int64) *Result {

	if len(p.Stages) == 0 {
		return &Result{Error: fmt.Errorf("No stages to be executed")}
	}

	var ticker *time.Ticker
	if p.expectedDuration != 0 && p.tick != 0 {
		// start progress update ticker
		ticker = time.NewTicker(p.tick)
		ctx, cancelProgress := context.WithCancel(context.Background())
		p.cancelProgress = cancelProgress
		go p.updateProgress(ticker, ctx)
	}

	buf, ok := buffersMap.get(p.Name)
	if !ok {
		return &Result{Error: fmt.Errorf("error creating output %s", p.Name)}
	}

	ctx, cancelDrain := context.WithCancel(context.Background())
	p.cancelDrain = cancelDrain
	go buf.drainBuffer(ctx)

	//defer buffersMap.remove(p.Name)
	defer p.waitForDrain()
	if p.expectedDuration != 0 && p.tick != 0 {
		defer ticker.Stop()
	}

	p.status(strconv.FormatInt(ID+1, 10) + " begin")

	defer p.status(strconv.FormatInt(ID+1, 10) + " end")

	//var request *Request
	//p.dispatch(request)
	request := &Request{Data: struct{ Order int64 }{Order: 1000 + ID}, KeyVal: map[string]interface{}{"Customer Order": 1000 + ID}}

	result := &Result{}
	for i, stage := range p.Stages {
		stage.index = i
		result = stage.run(request)
		if result.Error != nil {
			p.status("stage: " + stage.Name + " failed !!! ")
			return result
		}
		request.Data = result.Data
		request.KeyVal = result.KeyVal
	}

	return result
}

// Run the pipeline. The stages are executed in sequence while steps may be concurrent or sequential.
func (p *Pipeline) RunWithReq(request *Request) *Result {

	if len(p.Stages) == 0 {
		return &Result{Error: fmt.Errorf("No stages to be executed")}
	}

	var ticker *time.Ticker
	if p.expectedDuration != 0 && p.tick != 0 {
		// start progress update ticker
		ticker = time.NewTicker(p.tick)
		ctx, cancelProgress := context.WithCancel(context.Background())
		p.cancelProgress = cancelProgress
		go p.updateProgress(ticker, ctx)
	}

	buf, ok := buffersMap.get(p.Name)
	if !ok {
		return &Result{Error: fmt.Errorf("error creating output %s", p.Name)}
	}

	ctx, cancelDrain := context.WithCancel(context.Background())
	p.cancelDrain = cancelDrain
	go buf.drainBuffer(ctx)

	defer buffersMap.remove(p.Name)
	defer p.waitForDrain()
	if p.expectedDuration != 0 && p.tick != 0 {
		defer ticker.Stop()
	}
	defer p.status("end")

	p.status("begin")
	//request := &Request{Data: struct{ Order int64 }{Order: 1000}, KeyVal: map[string]interface{}{"Customer Order": 1000}}
	//var request *Request
	//p.dispatch(request)
	result := &Result{}
	for i, stage := range p.Stages {
		stage.index = i
		result = stage.run(request)
		if result.Error != nil {
			p.status("stage: " + stage.Name + " failed !!! ")
			return result
		}
		request.Data = result.Data
		request.KeyVal = result.KeyVal
	}

	return result
}

// Run the pipeline. The stages are executed in sequence while steps may be concurrent or sequential.
func (p *Pipeline) Run() *Result {

	if len(p.Stages) == 0 {
		return &Result{Error: fmt.Errorf("No stages to be executed")}
	}

	var ticker *time.Ticker
	if p.expectedDuration != 0 && p.tick != 0 {
		// start progress update ticker
		ticker = time.NewTicker(p.tick)
		ctx, cancelProgress := context.WithCancel(context.Background())
		p.cancelProgress = cancelProgress
		go p.updateProgress(ticker, ctx)
	}

	buf, ok := buffersMap.get(p.Name)
	if !ok {
		return &Result{Error: fmt.Errorf("error creating output %s", p.Name)}
	}

	ctx, cancelDrain := context.WithCancel(context.Background())
	p.cancelDrain = cancelDrain
	go buf.drainBuffer(ctx)

	defer buffersMap.remove(p.Name)
	defer p.waitForDrain()
	if p.expectedDuration != 0 && p.tick != 0 {
		defer ticker.Stop()
	}
	defer p.status("end")

	p.status("begin")

	request := &Request{Data: struct{ Order int64 }{Order: int64(1000)}, KeyVal: map[string]interface{}{"Customer Order": 1000}}
	//var request *Request
	//p.dispatch(request)
	result := &Result{}
	for i, stage := range p.Stages {
		stage.index = i
		result = stage.run(request)
		if result.Error != nil {
			p.status("stage: " + stage.Name + " failed !!! ")
			return result
		}
		request.Data = result.Data
		request.KeyVal = result.KeyVal
	}

	return result
}

// Out collects the status output from the stages and steps
func (p *Pipeline) Out() (<-chan string, error) {
	// add a new listener
	out := make(chan string, p.outbufferlen)
	err := buffersMap.appendOutBuffer(p.Name, out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// GetDuration returns the current time spent by the pipleline
func (p *Pipeline) GetDuration() time.Duration {
	return p.duration
}

// GetProgressPercent of the pipeline
func (p *Pipeline) GetProgressPercent() (<-chan int64, error) {
	pg := make(chan int64, 1)
	err := buffersMap.appendProgressBuffer(p.Name, pg)
	if err != nil {
		return nil, err
	}
	return pg, nil
}

// started as a goroutine
func (p *Pipeline) updateProgress(ticker *time.Ticker, ctx context.Context) {
	start := time.Now()
	for range ticker.C {
		p.duration = time.Since(start)
		percentDone := int64((p.duration.Seconds() / p.expectedDuration.Seconds()) * 100)
		// if estimate is incorrect don't overflow progress end
		if percentDone > 100 {
			percentDone = 99
		}
		buf, ok := buffersMap.get(p.Name)
		if !ok {
			return
		}
	loop:
		for _, pg := range buf.progress {
			select {
			case <-ctx.Done():
				break loop
			case pg <- percentDone:
			default:
				<-pg
				pg <- percentDone
			}
		}

	}
}

func (p *Pipeline) waitForDrain() {
	buf, ok := buffersMap.get(p.Name)
	if !ok {
		return
	}

	var empty = func() chan bool {
		emptyChan := make(chan bool)
		go func() {
			if len(buf.out) == 0 {
				return
			}

			pending := 0
			for _, o := range buf.out {
				pending += len(o)
			}

			if pending == 0 && len(buf.in) == 0 {
				emptyChan <- true
				return
			}

			emptyChan <- false
		}()
		return emptyChan
	}

loop:
	for {
		select {
		case empty := <-empty():
			if empty {
				break loop
			}
		case <-time.After(p.DrainTimeout):
			break loop
		}
	}

	p.cancelDrain()
	if p.cancelProgress != nil {
		p.cancelProgress()
	}
}

// status writes a line to the out channel
func (p *Pipeline) status(line string) {
	red := color.New(color.FgRed).SprintFunc()
	line = red("[pipeline]") + "[" + p.Name + "]: " + line
	send(p.Name, line)
}

func spaceMap(str string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, str)
}

func (p *Pipeline) Submit(job Job) {
	p.WorkChan <- job
}

func (p *Pipeline) process() {
	for {
		select {
		case job := <-p.WorkChan: // listen to any submitted job on the WorkChan
			// wait for a worker to submit JobChan to JobQueue
			// note that this JobQueue is shared among all workers.
			// Whenever there is an available JobChan on JobQueue pull it
			jobChan := <-p.Queue

			// Once a jobChan is available, send the submitted Job on this JobChan
			jobChan <- job
		}
	}
}

func (p *Pipeline) dispatch(request *Request) {

	go func() {
		for {
			// when available, put the JobChan again on the JobPool
			// and wait to receive a job
			p.Queue <- p.WorkChan
			select {
			case job := <-p.WorkChan:
				request = &Request{Data: struct{ Order int64 }{Order: job.JobID()}, KeyVal: map[string]interface{}{"Customer Order": job.JobID()}}
				//return
				// when a job is received, process
				//callApi(job.ID, wr.ID, c)
				//job.Dojob(wr.ID, job)
				//dojob(wr.ID, job)
				//case <-p.Quit:
				//	// a signal on this channel means someone triggered
				//	// a shutdown for this worker
				//	close(p.WorkChan)
				//	return
			}
		}
	}()

	go p.process()
}
