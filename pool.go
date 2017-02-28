package crawler

const maxThread = 100

type executorPool struct {
	Max     int
	Tickets chan *struct{}
}

func newExecutorPool() *executorPool {
	p := &executorPool{}
	p.Max = maxThread // setting

	p.Tickets = make(chan *struct{}, p.Max)
	for i := 0; i < p.Max; i++ {
		p.Tickets <- &struct{}{}
	}

	return p
}

func (p *executorPool) Return(ticket *struct{}) {
	if ticket == nil {
		return
	}
	p.Tickets <- ticket
}

func (p *executorPool) ActiveCount() int {
	return p.Max - len(p.Tickets)
}
