package logical

import (
	"context"
	gological "github.com/go-various/goplugin/logical"
	"github.com/go-various/pool"
)

type WorkerReply struct {
	Err    error               `json:"err"`
	Result *gological.Response `json:"result"`
}
type reader struct {
	c chan<- *WorkerReply
}

func NewReader(c chan<- *WorkerReply) *reader {
	return &reader{c}
}

func (r *reader) Update(result interface{}, err error) {
	if nil == result {
		r.c <- &WorkerReply{Err: err, Result: nil}
		return
	}
	r.c <- &WorkerReply{Err: err, Result: result.(*gological.Response)}
}

func (m *Transport) NewObserver(c chan<- *WorkerReply) pool.Observer {
	return NewReader(c)
}

func (m *Transport) StartWorkerPool() {
	poolLogger := m.Logger.Named("pool-0")
	m.WorkerPool = pool.NewWorkerPool("pool-0", context.Background(), poolLogger)

	for i := 0; i < m.workerSize; i++ {
		m.WorkerPool.NewWorker(m.backend)
	}

	m.WorkerPool.StartWorkers()
	go m.WorkerPool.Start()

}
