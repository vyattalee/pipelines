package worker1

import (
	"time"
)

type Pipeline struct {
	ID   int
	Name string
	//Machines  []*Machine
	Machines                chan Machine
	PipelineDone            chan struct{}
	SemiFinishedProductList chan SemiFinishedProduct
	CreatedAt               time.Time
	UpdatedAt               time.Time
}

func NewPipeline(ID int, Name string) *Pipeline {
	return &Pipeline{
		ID:   ID,
		Name: Name,
		//Machines:  make([]*Machine, 2),
		Machines:                make(chan Machine, 2),
		PipelineDone:            make(chan struct{}),
		SemiFinishedProductList: make(chan SemiFinishedProduct, 10),
		CreatedAt:               time.Now(),
		UpdatedAt:               time.Now(),
	}
}
