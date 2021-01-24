package logical

import (
	"context"
	"github.com/go-various/pool"
)

type WorkerReply struct {
	Err    error
	Result interface{}
}
type reader struct {
	c chan<- *WorkerReply
}

func NewReader(c chan<- *WorkerReply) *reader {
	return &reader{c}
}

func (r *reader) Update(result interface{}, err error) {
	r.c <- &WorkerReply{Err: err, Result: result}
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
