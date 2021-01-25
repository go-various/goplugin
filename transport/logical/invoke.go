package logical

import (
	"github.com/go-various/goplugin/logical"
	"github.com/go-various/pool"
)

func (m *Transport) Invoke(backend string, request *logical.Request) *WorkerReply {

	workerData := &WorkerData{
		Backend: backend,
		Request: request,
	}

	output := make(chan *WorkerReply, 1)
	defer close(output)

	subject := pool.NewSubject(workerData)

	subject.Observer(m.NewObserver(output))

	m.WorkerPool.Input(subject)

	select {
	case d := <-output:
		return d
	}

}
