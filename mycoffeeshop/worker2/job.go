package worker2

import "time"

type JobInterface interface {
	JobID() int64
	JobName() string
}

type Job struct {
	ID   int64
	Name string
	//Dojob     func(id int, job Job)
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (j Job) JobID() int64 {
	return j.ID
}

func (j Job) JobName() string {
	return j.Name
}

type SemiFinishedProduct struct {
	ProductId          int64
	ProductDescription string
}

func (s SemiFinishedProduct) JobID() int64 {
	return s.ProductId
}

func (s SemiFinishedProduct) JobName() string {
	return s.ProductDescription
}
