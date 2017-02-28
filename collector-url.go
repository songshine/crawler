package crawler

import (
	"log"

	"github.com/songshine/crawler/ruler"
)

type URLCollector interface {
	Next() (string, bool)
}

type urlCollectorImp struct {
	pager   ListPager
	ruler   ruler.Interface
	urlChan chan string
}

func NewURLCollector(p ListPager, r ruler.Interface) URLCollector {
	c := &urlCollectorImp{
		pager:   p,
		ruler:   r,
		urlChan: make(chan string, 500),
	}
	go func() {
		for {
			p, done := p.Next()
			if done {
				break
			}
			all := r.Get(p, true)
			for _, a := range all {
				c.urlChan <- a
			}
		}
		close(c.urlChan)
	}()
	return c
}

func (c *urlCollectorImp) Next() (url string, done bool) {
	log.Printf(">>> There are still %d urls needed to process\n", len(c.urlChan))

	url, ok := <-c.urlChan
	return url, !ok
}
