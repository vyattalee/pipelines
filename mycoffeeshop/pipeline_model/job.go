package pipeline

import "time"

type Job interface {
	JobID() int64
	JobName() string
}

type CustomerOrder struct {
	ID   int64
	Name string
	//Dojob     func(id int, job Job)
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (c CustomerOrder) JobID() int64 {
	return c.ID
}

func (c CustomerOrder) JobName() string {
	return c.Name
}

type SemiFinishedProduct struct {
	ProductId          int64
	ProductDescription string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

func (s SemiFinishedProduct) JobID() int64 {
	return s.ProductId
}

func (s SemiFinishedProduct) JobName() string {
	return s.ProductDescription
}
