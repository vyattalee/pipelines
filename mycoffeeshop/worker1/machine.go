package worker1

import (
	"fmt"
	"log"
	"time"
)

type Machine interface {
	dojob(worker Worker, job JobInterface) SemiFinishedProduct
	name() string
}

var grindBeanTime, espressoCoffeeTime, steamMilkTime time.Duration = 1, 2, 3

type GrindBeanMachine struct {
}

func (g *GrindBeanMachine) dojob(worker Worker, job JobInterface) SemiFinishedProduct {
	start := time.Now()
	prefix := fmt.Sprintf("Worker[%d]-Job[%d::%s]", worker.ID, job.JobID(), job.JobName())
	fmt.Println(prefix, "start to do job!")
	time.Sleep(time.Millisecond * grindBeanTime)
	end := time.Now()
	log.Print(end.Sub(start).Seconds())
	return SemiFinishedProduct{(job.JobID()), "grindBeanSemiFinishedProduct"}
}

func (g *GrindBeanMachine) name() string {
	return "GrindBeanMachine"
}

type EspressoCoffeeMachine struct {
}

func (e *EspressoCoffeeMachine) dojob(worker Worker, job JobInterface) SemiFinishedProduct {
	start := time.Now()
	prefix := fmt.Sprintf("Worker[%d]-Job[%d::%s]", worker.ID, job.JobID(), job.JobName())
	fmt.Println(prefix, "start to do job!")
	time.Sleep(time.Millisecond * espressoCoffeeTime)
	end := time.Now()
	log.Print(end.Sub(start).Seconds())
	return SemiFinishedProduct{(job.JobID()), "espressoCoffeeSemiFinishedProduct"}
}

func (e *EspressoCoffeeMachine) name() string {
	return "EspressoCoffeeMachine"
}

type SteamMilkMachine struct {
}

func (s *SteamMilkMachine) dojob(worker Worker, job JobInterface) SemiFinishedProduct {
	start := time.Now()
	prefix := fmt.Sprintf("Worker[%d]-Job[%d::%s]", worker.ID, job.JobID(), job.JobName())
	fmt.Println(prefix, "start to do job!")
	time.Sleep(time.Millisecond * steamMilkTime)
	end := time.Now()
	log.Print(end.Sub(start).Seconds())
	return SemiFinishedProduct{(job.JobID()), "steamMilkSemiFinishedProduct"}
}

func (s *SteamMilkMachine) name() string {
	return "SteamMilkMachine"
}
