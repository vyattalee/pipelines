package worker2

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
	worker.Name = g.name()
	prefix := fmt.Sprintf("Worker[%d::%s]", worker.ID, worker.Name)
	postfix := fmt.Sprintf("Job[%d::%s]", job.JobID(), job.JobName())
	fmt.Println(prefix, "start to do", postfix, "!")
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
	worker.Name = e.name()
	prefix := fmt.Sprintf("Worker[%d::%s]", worker.ID, worker.Name)
	postfix := fmt.Sprintf("Job[%d::%s]", job.JobID(), job.JobName())
	fmt.Println(prefix, "start to do", postfix, "!")
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
	worker.Name = s.name()
	prefix := fmt.Sprintf("Worker[%d::%s]", worker.ID, worker.Name)
	postfix := fmt.Sprintf("Job[%d::%s]", job.JobID(), job.JobName())
	fmt.Println(prefix, "start to do", postfix, "!")
	time.Sleep(time.Millisecond * steamMilkTime)
	end := time.Now()
	log.Print(end.Sub(start).Seconds())
	return SemiFinishedProduct{(job.JobID()), "steamMilkSemiFinishedProduct"}
}

func (s *SteamMilkMachine) name() string {
	return "SteamMilkMachine"
}

type HahaMachine struct {
}

func (s *HahaMachine) dojob(worker Worker, job JobInterface) SemiFinishedProduct {
	start := time.Now()
	worker.Name = s.name()
	prefix := fmt.Sprintf("Worker[%d::%s]", worker.ID, worker.Name)
	postfix := fmt.Sprintf("Job[%d::%s]", job.JobID(), job.JobName())
	fmt.Println(prefix, "start to do", postfix, "!")
	time.Sleep(time.Millisecond * steamMilkTime)
	end := time.Now()
	log.Print(end.Sub(start).Seconds())
	return SemiFinishedProduct{(job.JobID()), "HahaSemiFinishedProduct"}
}

func (s *HahaMachine) name() string {
	return "HahaMachine"
}
