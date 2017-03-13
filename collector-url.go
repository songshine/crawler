package crawler

import (
	"log"

	"github.com/songshine/crawler/ruler"
)

// URLCollector defines a collector to collect URLs, and invoking Next to
// loop all URLs. When there are no more URLs available, Next returns
// true for the second value returned.
type URLCollector interface {
	Next() (string, bool)
}

type urlCollectorImp struct {
	pager   ListPager
	ruler   ruler.Interface
	urlChan chan string
}

// NewURLCollector creates a URLCollector implementation.
func NewURLCollector(p ListPager, r ruler.Interface) URLCollector {
	c := &urlCollectorImp{
		pager:   p,
		ruler:   r,
		urlChan: make(chan string, 50),
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
	log.Printf(">>> URL count in collector: %d \n", len(c.urlChan))

	url, ok := <-c.urlChan
	return url, !ok
}
