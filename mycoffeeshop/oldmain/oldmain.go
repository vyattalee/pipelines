// The coffee command simulates a small parallel pipeline and outputs CSV.
//
// The pipeline consists of three stages: grinding coffee beans,
// preparing espresso, and steaming milk.  Each stage contends on the
// respective machine (grinder, espresso machine, steamer).
//
// This simulation reports throughput, latency, and utilization.
// It can also create an execution trace with the --trace flag.
package oldmain

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"runtime"
	"runtime/trace"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	mode = flag.String("mode", "ideal", `comma-separated list of modes:
ideal: no synchronization, no contention overhead.  Fails the race detector.
locking: one lock, maximal contention.
finelocking: one lock per machine, permitting greater parallelism.
parsteam: finelocking with steaming happening in parallel with the other stages.
americano: skip the steamMilk stage, but still makeLatte to add the water.
espresso: skip the steamMilk and makeLatte stages.
linearpipe-N: a pipeline with one goroutine per machine.
splitpipe-N: a pipeline with the steamer stage happening in parallel with the other stages.
multi-N: finelocking but with N copies of each machine.
multipipe-N: N copies of linearpipe.
`)
	duration  = flag.Duration("dur", 1*time.Second, "perf test duration")
	interval  = flag.Duration("interval", 0, "perf test request interval")
	grindTime = flag.Duration("grind", 1*time.Millisecond, "grind phase duration")
	pressTime = flag.Duration("press", 1*time.Millisecond, "press phase duration")
	steamTime = flag.Duration("steam", 1*time.Millisecond, "steam phase duration")
	latteTime = flag.Duration("latte", 1*time.Millisecond, "latte phase duration")
	jitter    = flag.Duration("jitter", 0, "add uniform random duration in [-jitter/2,+jitter/2] to each phase")
	printDurs = flag.Bool("printdurs", false, "print duration distribution of each phase")
	traceFlag = flag.String("trace", "", "execution trace file, e.g., ./trace.out")
	header    = flag.Bool("header", true, "whether to print CSV header")
	pars      intList
	maxqs     intList
)

func init() {
	flag.Var(&pars, "par", "comma-separated list of perf test parallelism (how many brews to run in parallel)")
	flag.Var(&maxqs, "maxq", "comma-separated max lengths of the request queue (how many calls to queue up)")
}

type intList []int

func (il *intList) Set(s string) error {
	ss := strings.Split(s, ",")
	for _, s := range ss {
		n, err := strconv.Atoi(s)
		if err != nil {
			return err
		}
		*il = append(*il, n)
	}
	return nil
}

func (il *intList) String() string {
	var ss []string
	for _, n := range *il {
		ss = append(ss, strconv.Itoa(n))
	}
	return strings.Join(ss, ",")
}

// A machine keeps a count of how often it's been used and a sample of
// latencies.
type machine struct {
	name string
	sync.Mutex
	*sampler
	number int64 //number of product made
}

func newMachine(name string) *machine {
	return &machine{
		name:    name,
		sampler: newSampler(),
	}
}

func (m *machine) close() {
	if m == nil {
		return
	}
	m.sampler.close()
	if *printDurs {
		log.Println(m.name, ":", m.sampler)
	}
}

// Shared state, requiring synchronization
var (
	grinder         = newMachine("grinder")
	espressoMachine = newMachine("presser")
	steamer         = newMachine("steamer")
)

func resetMachines() {
	grinder.close()
	espressoMachine.close()
	steamer.close()
	grinder = newMachine("grinder")
	espressoMachine = newMachine("presser")
	steamer = newMachine("steamer")
}

// Named types for the pipeline elements.
type (
	grounds int
	coffee  int
	milk    int
	latte   int
)

// Ideal case: no contention (fails the race detector with par > 1)
func idealBrew() latte {
	grounds := grindCoffee(grinder)
	coffee := makeEspresso(espressoMachine, grounds)
	milk := steamMilk(steamer)
	return makeLatte(coffee, milk)
}

// Simulate one millisecond of prep time.
// Each stage adds a latency sample to the provided machine.
// Synchronization must happen outside these stages.

func runPhase(d, jitter time.Duration) time.Duration {
	if jitter > 0 {
		d += time.Duration(rand.Int63n((jitter).Nanoseconds()))
		d -= jitter / 2
	}
	start := time.Now()
	useCPU(d)
	return time.Since(start)
}

func grindCoffee(grinder *machine) grounds {
	grinder.add(runPhase(*grindTime, *jitter))
	return grounds(0)
}

func makeEspresso(espressoMachine *machine, grounds grounds) coffee {
	espressoMachine.add(runPhase(*pressTime, *jitter))
	return coffee(grounds)
}

func steamMilk(steamer *machine) milk {
	steamer.add(runPhase(*steamTime, *jitter))
	return milk(0)
}

func makeLatte(coffee coffee, milk milk) latte {
	// No shared state to contend on.
	runPhase(*latteTime, 0) // no jitter
	return latte(int(coffee) + int(milk))
}

// Locking case: contention on the whole kitchen.
var kitchen sync.Mutex

func lockingBrew() latte {
	kitchen.Lock()
	defer kitchen.Unlock()
	grounds := grindCoffee(grinder)
	coffee := makeEspresso(espressoMachine, grounds)
	milk := steamMilk(steamer)
	return makeLatte(coffee, milk)
}

// Fine-grain locking reduces contention.

func fineLockingBrew() latte {
	grounds := lockingGrind()
	coffee := lockingPress(grounds)
	milk := lockingSteam()
	return makeLatte(coffee, milk)
}

func lockingGrind() grounds {
	grinder.Lock()
	defer grinder.Unlock()
	return grindCoffee(grinder)
}

func lockingPress(grounds grounds) coffee {
	espressoMachine.Lock()
	defer espressoMachine.Unlock()
	return makeEspresso(espressoMachine, grounds)
}

func lockingSteam() milk {
	steamer.Lock()
	defer steamer.Unlock()
	return steamMilk(steamer)
}

// Paralellizing operations can help, provided there's available CPU.
// Can steam milk while grinding & pressing, but this loses to
// fine-grain locking when all CPUs utilized.
func parallelSteaming() latte {
	c := make(chan milk, 1)
	go func() {
		c <- lockingSteam()
	}()
	grounds := lockingGrind()
	coffee := lockingPress(grounds)
	milk := <-c
	return makeLatte(coffee, milk)
}

// Americano skips the steamMilk stage.  This simulates making an RPC or doing a
// cache lookup instead of burning CPU for that stage.  This wins over
// fine-grain locking.
func americano() latte {
	grounds := lockingGrind()
	coffee := lockingPress(grounds)
	water := milk(0)
	return makeLatte(coffee, water)
}

// Espresso skips the steamMilk and makeLatte stages.  This shows the benefit of
// skipping optional work (and possibly delivering degraded results).
func espresso() latte {
	grounds := lockingGrind()
	coffee := lockingPress(grounds)
	return latte(coffee) // no milk or water
}

// Multiple machines reduce contention.
var grinders, espressoMachines, steamers chan *machine

// newMachines returns a channel containing n machines in its buffer.
func newMachines(name string, n int) chan *machine {
	c := make(chan *machine, n)
	for i := 0; i < n; i++ {
		c <- newMachine(fmt.Sprintf("%s%d", name, i))
	}
	return c
}

func closeMachines(c chan *machine) {
	close(c)
	for m := range c {
		m.close()
	}
}

func multiBrew() latte {
	grounds := multiGrind()
	coffee := multiPress(grounds)
	milk := multiSteam()
	return makeLatte(coffee, milk)
}

func multiGrind() grounds {
	m := <-grinders
	grounds := grindCoffee(m)
	grinders <- m
	return grounds
}

func multiPress(grounds grounds) coffee {
	m := <-espressoMachines
	coffee := makeEspresso(m, grounds)
	espressoMachines <- m
	return coffee
}

func multiSteam() milk {
	m := <-steamers
	milk := steamMilk(m)
	steamers <- m
	return milk
}

// Linear pipeline
type order struct {
	grounds grounds
	coffee  coffee
	milk    chan milk
}

type linearPipeline struct {
	grinderMachine  *machine
	espressoMachine *machine
	steamerMachine  *machine

	orders            chan order
	ordersWithGrounds chan order
	ordersWithCoffee  chan order
	done              chan int
}

func newLinearPipeline(buffer int) *linearPipeline {
	p := &linearPipeline{
		grinderMachine:    newMachine("grinder"),
		espressoMachine:   newMachine("presser"),
		steamerMachine:    newMachine("steamer"),
		orders:            make(chan order, buffer),
		ordersWithGrounds: make(chan order, buffer),
		ordersWithCoffee:  make(chan order, buffer),
		done:              make(chan int),
	}
	go p.grinder()
	go p.presser()
	go p.steamer()
	return p
}

// newLinearPipelineMulti returns a pipeline that uses a shared orders channel
// from multiPipeline.
func newLinearPipelineMulti(i int, orders chan order) *linearPipeline {
	p := &linearPipeline{
		grinderMachine:    newMachine(fmt.Sprintf("grinder%d", i)),
		espressoMachine:   newMachine(fmt.Sprintf("presser%d", i)),
		steamerMachine:    newMachine(fmt.Sprintf("steamer%d", i)),
		orders:            orders,
		ordersWithGrounds: make(chan order, 10), // small buffer
		ordersWithCoffee:  make(chan order, 10),
		done:              make(chan int),
	}
	go p.grinder()
	go p.presser()
	go p.steamer()
	return p
}

func (p *linearPipeline) brew() latte {
	// Buffer result channel to prevent deadlock.
	o := order{milk: make(chan milk, 1)}
	p.orders <- o
	milk := <-o.milk
	return makeLatte(o.coffee, milk)
}

func (p *linearPipeline) grinder() {
	for o := range p.orders {
		o.grounds = grindCoffee(p.grinderMachine)
		p.ordersWithGrounds <- o
	}
	close(p.ordersWithGrounds)
}

func (p *linearPipeline) presser() {
	for o := range p.ordersWithGrounds {
		o.coffee = makeEspresso(p.espressoMachine, o.grounds)
		p.ordersWithCoffee <- o
	}
	close(p.ordersWithCoffee)
}

func (p *linearPipeline) steamer() {
	for o := range p.ordersWithCoffee {
		o.milk <- steamMilk(p.steamerMachine)
	}
	close(p.done)
}

func (p *linearPipeline) close() {
	close(p.orders)
	<-p.done
	p.grinderMachine.close()
	p.espressoMachine.close()
	p.steamerMachine.close()
}

// Americano pipeline skips the steamMilk step.
func newAmericanoPipeline(buffer int) *linearPipeline {
	p := &linearPipeline{
		grinderMachine:    newMachine("grinder"),
		espressoMachine:   newMachine("presser"),
		orders:            make(chan order, buffer),
		ordersWithGrounds: make(chan order, buffer),
		done:              make(chan int),
	}
	go p.grinder()
	go p.americanoPresser()
	return p
}

func (p *linearPipeline) americanoBrew() latte {
	// Buffer result channel to prevent deadlock.
	o := order{milk: make(chan milk, 1)}
	p.orders <- o
	water := <-o.milk
	return makeLatte(o.coffee, water)
}

func (p *linearPipeline) americanoPresser() {
	for o := range p.ordersWithGrounds {
		o.coffee = makeEspresso(p.espressoMachine, o.grounds)
		o.milk <- milk(0) // water
	}
	close(p.done)
}

// Espresso pipeline skips the steamMilk and makeLatte steps.
func newEspressoPipeline(buffer int) *linearPipeline {
	p := &linearPipeline{
		grinderMachine:    newMachine("grinder"),
		espressoMachine:   newMachine("presser"),
		orders:            make(chan order, buffer),
		ordersWithGrounds: make(chan order, buffer),
		done:              make(chan int),
	}
	go p.grinder()
	go p.americanoPresser()
	return p
}

func (p *linearPipeline) espressoBrew() latte {
	// Buffer result channel to prevent deadlock.
	o := order{milk: make(chan milk, 1)}
	p.orders <- o
	<-o.milk               // espresso done
	return latte(o.coffee) // no milk or water
}

// Split pipeline
type splitOrder struct {
	grounds grounds
	coffee  chan coffee
	milk    chan milk
}
type splitPipeline struct {
	grinderMachine  *machine
	espressoMachine *machine
	steamerMachine  *machine

	coffeeOrders      chan splitOrder
	ordersWithGrounds chan splitOrder
	presserDone       chan int

	milkOrders  chan splitOrder
	steamerDone chan int
}

func newSplitPipeline(buffer int) *splitPipeline {
	p := &splitPipeline{
		grinderMachine:    newMachine("grinder"),
		espressoMachine:   newMachine("presser"),
		steamerMachine:    newMachine("steamer"),
		coffeeOrders:      make(chan splitOrder, buffer),
		ordersWithGrounds: make(chan splitOrder, buffer),
		presserDone:       make(chan int),
		milkOrders:        make(chan splitOrder, buffer),
		steamerDone:       make(chan int),
	}
	go p.grinder()
	go p.presser()
	go p.steamer()
	return p
}

func (p *splitPipeline) brew() latte {
	o := splitOrder{
		// Buffer result channel to prevent deadlocks.
		coffee: make(chan coffee, 1),
		milk:   make(chan milk, 1),
	}
	p.coffeeOrders <- o
	p.milkOrders <- o
	coffee := <-o.coffee
	milk := <-o.milk
	return makeLatte(coffee, milk)
}

func (p *splitPipeline) grinder() {
	for o := range p.coffeeOrders {
		o.grounds = grindCoffee(p.grinderMachine)
		p.ordersWithGrounds <- o
	}
	close(p.ordersWithGrounds)
}

func (p *splitPipeline) presser() {
	for o := range p.ordersWithGrounds {
		o.coffee <- makeEspresso(p.espressoMachine, o.grounds)
	}
	close(p.presserDone)
}

func (p *splitPipeline) steamer() {
	for o := range p.milkOrders {
		o.milk <- steamMilk(p.steamerMachine)
	}
	close(p.steamerDone)
}

func (p *splitPipeline) close() {
	close(p.coffeeOrders)
	<-p.presserDone
	close(p.milkOrders)
	<-p.steamerDone
	p.grinderMachine.close()
	p.espressoMachine.close()
	p.steamerMachine.close()
}

// Multiple copies of linearPipeline, like multiple coffee shops.
type multiPipeline struct {
	orders chan order
	pipes  chan *linearPipeline
}

func newMultiPipeline(n int) *multiPipeline {
	p := &multiPipeline{
		orders: make(chan order),
		pipes:  make(chan *linearPipeline, n),
	}
	for i := 0; i < n; i++ {
		p.pipes <- newLinearPipelineMulti(i, p.orders)
	}
	return p
}

func (p *multiPipeline) brew() latte {
	lp := <-p.pipes
	o := order{milk: make(chan milk, 1)}
	lp.orders <- o
	p.pipes <- lp    // release the pipeline for other brew calls
	milk := <-o.milk // THEN wait for order to complete
	return makeLatte(o.coffee, milk)
}

func (p *multiPipeline) close() {
	close(p.orders)
	close(p.pipes)
	for lp := range p.pipes {
		<-lp.done
		lp.grinderMachine.close()
		lp.espressoMachine.close()
		lp.steamerMachine.close()
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())
	start := time.Now()
	log.Print("GOMAXPROCS=", runtime.GOMAXPROCS(0))
	flag.Parse()
	if len(pars) == 0 {
		pars = []int{1}
	}
	if len(maxqs) == 0 {
		maxqs = []int{0}
	}
	modes := strings.Split(*mode, ",")
	if len(modes) == 0 {
		modes = []string{"ideal"}
	}
	if *traceFlag != "" {
		traceFile, err := os.Create(*traceFlag)
		if err != nil {
			panic(err)
		}
		trace.Start(traceFile)
		defer func() {
			trace.Stop()
			if err := traceFile.Close(); err != nil {
				log.Panic(err)
			}
		}()
	}
	// Run all combinations of modes, parallelisms, and maxqs.
	// Print output as CSV.
	if *header {
		fmt.Println(perfArgHeader + "," + perfResultHeader + ",jitter")
	}
	for _, mode := range modes {
		f, close := modeFunc(mode)
		for _, par := range pars {
			if par == 0 {
				par = runtime.GOMAXPROCS(0)
			}
			for _, maxq := range maxqs {
				arg := perfArg{
					mode:     mode,
					par:      par,
					maxq:     maxq,
					dur:      *duration,
					interval: *interval,
				}
				res := perfTest(arg, func() { f() })
				fmt.Println(arg.String() + "," + res.String() + "," + (*jitter).String())
			}
		}
		if close != nil {
			close()
		}
		fmt.Println(mode, "模式下共用时间：", time.Since(start))
	}
}

func modeFunc(mode string) (func() latte, func()) {
	var n int
	switch {
	case mode == "ideal":
		return idealBrew, resetMachines
	case mode == "locking":
		return lockingBrew, resetMachines
	case modeParam(mode, "multi-", &n):
		grinders = newMachines("grinder", n)
		espressoMachines = newMachines("presser", n)
		steamers = newMachines("steamer", n)
		return multiBrew, func() {
			closeMachines(grinders)
			grinders = nil
			closeMachines(espressoMachines)
			espressoMachines = nil
			closeMachines(steamers)
			steamers = nil
		}
	case mode == "finelocking":
		return fineLockingBrew, resetMachines
	case mode == "parsteam":
		return parallelSteaming, resetMachines
	case mode == "americano":
		return americano, resetMachines
	case mode == "espresso":
		return espresso, resetMachines
	case modeParam(mode, "linearpipe-", &n):
		p := newLinearPipeline(n)
		return p.brew, p.close
	case modeParam(mode, "americanopipe-", &n):
		p := newAmericanoPipeline(n)
		return p.americanoBrew, p.close
	case modeParam(mode, "espressopipe-", &n):
		p := newEspressoPipeline(n)
		return p.espressoBrew, p.close
	case modeParam(mode, "splitpipe-", &n):
		p := newSplitPipeline(n)
		return p.brew, p.close
	case modeParam(mode, "multipipe-", &n):
		p := newMultiPipeline(n)
		return p.brew, p.close
	case modeParam(mode, "myLattepipe-", &n):
		p := newMyLattePipeline(n)
		return p.latte, p.closeP3
	}
	log.Panicf("unknown mode: %s", mode)
	return nil, nil
}

func modeParam(mode, prefix string, n *int) bool {
	if !strings.HasPrefix(mode, prefix) {
		return false
	}
	var err error
	*n, err = strconv.Atoi((mode)[len(prefix):])
	if err != nil {
		log.Panicf("bad mode %s: %v", mode, err)
	}
	return true
}

//////////////////////////////////////////////////////////////////////////////////
/////////design pattern
//////////////////////////////////////////////////////////////////////////////////
type Machine interface {
	running() SemiFinishedProduct
}

type SemiFinishedProduct struct {
	productId          int64
	productDescription string
}

var grindBeanTime, espressoCoffeeTime, steamMilkTime time.Duration = 1, 2, 3

type grindBeanMachine struct {
	order chan Order
}

func (g *grindBeanMachine) running() SemiFinishedProduct {

	time.Sleep(time.Millisecond * grindBeanTime)
	return SemiFinishedProduct{}
}

type espressoCoffeeMachine struct {
}

func (e *espressoCoffeeMachine) running() SemiFinishedProduct {
	time.Sleep(time.Millisecond * espressoCoffeeTime)
	return SemiFinishedProduct{}
}

type steamMilkMachine struct {
}

func (e *steamMilkMachine) running() SemiFinishedProduct {
	time.Sleep(time.Millisecond * steamMilkTime)
	return SemiFinishedProduct{}
}

func worker(machines chan Machine, semiFinProduct chan SemiFinishedProduct) {
	for machine := range machines {

		semiFinProduct <- machine.running()

		//select{
		//case <- order:
		//	semiFinProduct <- machine.running()
		//	case <- chanGround:
		//		semiFinProduct <- machine.running()
		//}
		//
	}
}

type dispatcher struct {
	order           chan Order
	pipelineMachine chan Machine //using for grindBeanMachine and espressoCoffeeMachine in one pipeline
}

//1. 接收到订单 <- order，
//1.1 启动grindBeanMachine.running，结束后开始espressoCoffeeMachine.running，控制在pipelineMachine中
//1.2 启动steamMilkMachine.running
//2. 同时接收到 <- chanEspressoCoffee和 <- chanSteamMilk，可以makeLatte(coffee, milk)

func pipeline() {

}

//////////////////////////////////////////////////////////////////////////////////

//order ---> grinds ---> coffee(espresso) ---> latte
//			 milk --->
//有订单order才会启动grinds和milk machine，并且grinds结束才能进行coffee制作，milk和coffe都有的情况下才能合并成latte

func newMyLattePipeline(buffer int) *MyLattePipeline {
	p := &MyLattePipeline{
		grinderMachine:  newMachine("grinder"),
		espressoMachine: newMachine("presser"),
		steamerMachine:  newMachine("steamer"),
		orders:          make(chan Order, buffer),
		chanGround:      make(chan int64, buffer),
		chanCoffee:      make(chan int64, buffer),
		chanMilk:        make(chan int64, buffer),
		chanLatte:       make(chan LatteRecord, buffer),
		chanCoffeePost:  make(chan PostRecord, buffer),
		chanMilkPost:    make(chan PostRecord, buffer),

		done: make(chan int),
	}

	//go p.dispatchOrder()
	//go p.coffeeBeanGrinder()
	//go p.espressoCoffee()
	//go p.milkSteamer()

	//go p.grinder()
	//go p.presser()
	//go p.steamer()
	//go p.processbychanGround()
	//go p.processbyorders()
	return p
}

type Order struct {
	orderId          int64
	orderDescription string
}

type PostRecord struct {
	orderId   int64
	recordNum int64
}

type LatteRecord struct {
	orderId   int64
	coffeeNum coffee
	milkNum   milk
}

// proposal 3
type MyLattePipeline struct {
	grinderMachine  *machine
	espressoMachine *machine
	steamerMachine  *machine

	orders         chan Order
	chanGround     chan int64
	chanCoffee     chan int64
	chanMilk       chan int64
	chanCoffeePost chan PostRecord
	chanMilkPost   chan PostRecord
	chanLatte      chan LatteRecord
	done           chan int
}

func (p *MyLattePipeline) makeLatteP3(cof coffee, mlk milk) latte {
	// No shared state to contend on.
	runPhase(*latteTime, 0) // no jitter
	return latte(int64(cof) + int64(mlk))
}
func (p *MyLattePipeline) latte() latte {

	order_init(p)

	grindingBean_steamingMilk_parallel_step(p)

	espressoCoffee_step(p)

	var coffeeNum, milkNum int64

	select {
	case coffeeNum = <-p.chanCoffee:
		log.Println("espresso coffee ready")
	case <-time.After(time.Second * 1):
		log.Println("espresso coffee timeout")
	}

	select {
	case milkNum = <-p.chanCoffee:
		log.Println("steam milk ready")
	case <-time.After(time.Second * 1):
		log.Println("steam milk timeout")
	}

	return p.makeLatteP3(coffee(coffeeNum), milk(milkNum))
}

func espressoCoffee_step(p *MyLattePipeline) {
	select {
	case ground := <-p.chanGround:
		go p.espressoCoffeeP3(grounds(ground))
	case <-time.After(time.Second * 1):
		log.Println("espressoCoffeeP3 timeout!")
	}
}

func grindingBean_steamingMilk_parallel_step(p *MyLattePipeline) {
	select {
	case order := <-p.orders:
		go p.grindingBeanP3(order.orderId)
		go p.steamingMilkP3(order.orderId)
	case <-time.After(time.Second * 1):
		log.Println("grindingBean & steaminMilk timeout!")

	}
}

func order_init(p *MyLattePipeline) {

	var o = []Order{
		//Order{1000, "first"}}
		Order{1000, "first"},
		Order{1001, "second"},
		Order{1002, "third"}}
	for i := range o {
		p.orders <- o[i]
	}
	close(p.orders)

}

func (p *MyLattePipeline) grindingBeanP3(orderId int64) grounds {
	runPhase(*grindTime, *jitter)
	p.chanGround <- (orderId)
	//defer close(p.chanGround)
	return grounds(orderId)
}
func (p *MyLattePipeline) steamingMilkP3(orderId int64) milk {
	runPhase(*steamTime, *jitter)
	p.chanMilk <- orderId
	//defer close(p.chanMilk)
	return milk(orderId)
}
func (p *MyLattePipeline) espressoCoffeeP3(ground grounds) coffee {
	runPhase(*pressTime, *jitter)
	p.chanCoffee <- int64(ground)
	//defer close(p.chanCoffee)
	return coffee(ground)
}
func (p *MyLattePipeline) closeP3() {
	close(p.chanGround)
	close(p.chanMilk)
	close(p.chanCoffee)
	close(p.chanCoffeePost)
	close(p.chanMilkPost)

}

// end of proposal 3

//proposal 2
//type MyLattePipeline struct {
//	grinderMachine  *machine
//	espressoMachine *machine
//	steamerMachine  *machine
//
//	orders         chan Order
//	chanGround     chan int64
//	chanCoffee     chan int64
//	chanMilk       chan int64
//	chanCoffeePost chan PostRecord
//	chanMilkPost   chan PostRecord
//	chanLatte      chan LatteRecord
//	done           chan int
//}

func (p *MyLattePipeline) dispatchOrder() {
	//for o := range p.orders {
	select {
	case order := <-p.orders:
		//p.chanGround <- grindCoffeeWithNum(p.grinderMachine)
		p.chanGround <- order.orderId
		p.chanMilk <- order.orderId
		log.Println("OrderId:", order.orderId, "entering beans grinder")
	case <-time.After(time.Second * 1):
		log.Println("grinder timeout")
	}

	//}

}

func (p *MyLattePipeline) coffeeBeanGrinder() {
	select {
	case id := <-p.chanGround:
		grindingCoffeeBean(p.grinderMachine)
		p.chanCoffee <- id
	case <-time.After(time.Second * 1):
		log.Println("coffeeBeanGrinder timeout")
	}
}

func grindingCoffeeBean(grinderMachine *machine) grounds {
	grinderMachine.add(runPhase(*grindTime, *jitter))
	grinderMachine.number += 1
	return grounds(grinderMachine.number)
}

func (p *MyLattePipeline) espressoCoffee() {
	select {
	case id := <-p.chanCoffee:
		p.chanCoffeePost <- PostRecord{id, int64(makingEspressoCoffee(p.espressoMachine))}
	case <-time.After(time.Second * 1):
		log.Println("espressoCoffee timeout")
	}
}

func makingEspressoCoffee(espressoMachine *machine) coffee {
	espressoMachine.add(runPhase(*pressTime, *jitter))
	espressoMachine.number += 1
	return coffee(espressoMachine.number)
}

func (p *MyLattePipeline) milkSteamer() {
	select {
	case id := <-p.chanMilk:
		p.chanCoffeePost <- PostRecord{id, int64(steamingMilk(p.steamerMachine))}
	case <-time.After(time.Second * 1):
		log.Println("coffeeBeanGrinder timeout")
	}
}

func steamingMilk(milkMachine *machine) milk {
	milkMachine.add(runPhase(*steamTime, *jitter))
	milkMachine.number += 1
	return milk(milkMachine.number)
}

func (p *MyLattePipeline) latteMaker() latte {

	var espressoRecord, milkRecord PostRecord

	select {
	case espressoRecord = <-p.chanCoffeePost:
		//p.chanLatte
		log.Println("espresso coffee record:", espressoRecord)
	case milkRecord = <-p.chanMilkPost:
		log.Println("milk steaming post record:", milkRecord)
	case <-time.After(time.Second * 1):
		log.Println("latteMaker timeout")
	}

	result := makeLatte(coffee(espressoRecord.recordNum), milk(milkRecord.recordNum))

	return result
}

func (p *MyLattePipeline) produce() latte {

	var o = []Order{
		//Order{1000, "first"}}
		Order{1000, "first"},
		Order{1001, "second"},
		Order{1002, "third"}}
	for i := range o {
		p.orders <- o[i]
	}

	//go p.dispatchOrder()
	//go p.coffeeBeanGrinder()
	//go p.espressoCoffee()
	//go p.milkSteamer()

	return p.latteMaker()

}

// end of proposal 2

//proposal 1

func (p *MyLattePipeline) processbyorders() {
	select {
	case order := <-p.orders:
		p.chanGround <- grindCoffeeWithNum(p.grinderMachine)
		p.chanMilk <- steamMilkWithNum(p.steamerMachine)
		log.Println("OrderId:", order, "entering processbyorders")
	case <-time.After(time.Second * 1):
		log.Println("processbyorders timeout")
	}
}

func (p *MyLattePipeline) processbychanGround() {
	select {
	case ground := <-p.chanGround:
		p.chanCoffee <- makeEspressoWithNum(p.espressoMachine, p.chanGround)
		log.Println("ground num:", ground, "entering processbychanGround espresso done")
	case <-time.After(time.Second * 1):
		log.Println("processbychanGround timeout")
	}
}

func grindCoffeeWithNum(grinder *machine) int64 {
	grinder.add(runPhase(*grindTime, *jitter))
	grinder.number += 1
	return grinder.number
}

func (p *MyLattePipeline) presser() {
	for o := range p.chanGround {
		p.chanCoffee <- makeEspressoWithNum(p.espressoMachine, p.chanGround)
		<-p.chanGround
		log.Println("OrderId:", o, "entering espresso presser")
	}
}

func makeEspressoWithNum(espressoMachine *machine, chanGround chan int64) int64 {
	espressoMachine.add(runPhase(*pressTime, *jitter))
	espressoMachine.number += 1
	return espressoMachine.number
}

func (p *MyLattePipeline) steamer() {
	for o := range p.orders {
		p.chanMilk <- steamMilkWithNum(p.steamerMachine)
		log.Println("OrderId:", o, "entering milk steamer")
	}
	//close(p.done)
}

func steamMilkWithNum(milk *machine) int64 {
	milk.add(runPhase(*grindTime, *jitter))
	milk.number += 1
	return milk.number
}

func (p *MyLattePipeline) brew() latte {
	//coffeeNum := <-p.chanCoffee
	//milkNum :=  <-p.chanMilk
	// Buffer result channel to prevent deadlock.
	//o := order{milk: make(chan milk, 1)}
	//p.orders <- o
	//milk := <-o.milk
	var o = []Order{
		Order{1000, "first"},
		Order{1001, "second"},
		Order{1002, "third"}}
	for i := range o {
		p.orders <- o[i]
	}

	var flag int = 0

	select {
	case <-p.chanCoffee:
		flag += 1
		log.Println("epresso make ready")
	case <-p.chanMilk:
		flag += 1
		log.Println("milk steam ready")

		if flag == 2 {

			log.Println("preparing make latte")
			makeLatteWithNum(p.done)

		}

	}
	result := makeLatte(0, 0)
	p.done <- 0
	return result
}

func makeLatteWithNum(done chan int) int64 {
	<-done
	log.Println("latte done!")
	//close(done)
	return 0

}

func (p *MyLattePipeline) close() {
	close(p.orders)
	close(p.chanGround)
	close(p.chanMilk)
	close(p.chanCoffee)
	close(p.chanLatte)
	close(p.chanCoffeePost)
	close(p.chanMilkPost)

}

// end of proposal 1
