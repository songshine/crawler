package pool

type Interface interface {
	Take() *struct{}
	Return(*struct{})
	ActiveCount() int
}

type executorPool struct {
	max     int
	tickets chan *struct{}
}

func New(max int) Interface {
	p := &executorPool{}
	p.max = max // setting

	p.tickets = make(chan *struct{}, p.max)
	for i := 0; i < p.max; i++ {
		p.tickets <- &struct{}{}
	}

	return p
}

func (p *executorPool) Take() *struct{} {
	return <-p.tickets
}

func (p *executorPool) Return(ticket *struct{}) {
	if ticket == nil {
		return
	}
	p.tickets <- ticket
}

func (p *executorPool) ActiveCount() int {
	return p.max - len(p.tickets)
}
