package crawler

import (
	"sync"

	"github.com/songshine/crawler/ruler"
)

const maxCache = 500

type (
	// FieldValues represents all useful data from a web page.
	FieldValues []string

	// FieldItem represents a rule how to crawl a field from a web page.
	FieldItem struct {
		Name    string
		FromURL bool
		Rule    ruler.Interface
	}

	fields []*FieldItem
)

// DataCollector represents a collector to crawl data from a web page.
type DataCollector interface {
	Collect() <-chan FieldValues
	Names() []string
	Add(string, ruler.Interface, bool)
}

// NewDataCollector creates a DataCollector instance.
func NewDataCollector(urlCollector URLCollector, items ...*FieldItem) DataCollector {
	return &dataCollectorImp{
		urlCollector: urlCollector,
		fields:       items,
		result:       make(chan FieldValues, maxCache),
		pool:         newExecutorPool(),
	}
}

func (f fields) Add(name string, rule ruler.Interface, fromURL bool) {
	f = append(f, &FieldItem{name, fromURL, rule})
}

func (f fields) Names() []string {
	ns := make([]string, 0, len(f))
	for _, i := range f {
		ns = append(ns, i.Name)
	}

	return ns
}

type dataCollectorImp struct {
	fields
	urlCollector URLCollector
	result       chan FieldValues
	pool         *executorPool
	wg           sync.WaitGroup
}

func (c *dataCollectorImp) Collect() <-chan FieldValues {
	go func() {
		for {
			url, done := c.urlCollector.Next()
			if done {
				break
			}
			c.wg.Add(1)
			ticket := <-c.pool.Tickets
			go c.colllecFromURL(url, ticket)
		}

		c.wg.Wait()
		close(c.result)
	}()
	return c.result
}

func (c *dataCollectorImp) colllecFromURL(url string, ticket *struct{}) {
	defer func() {
		c.pool.Return(ticket)
		c.wg.Done()
	}()

	resp, err := Get(url)
	if err != nil {
		return
	}

	vals := make(FieldValues, 0, len(c.fields))
	for _, i := range c.fields {
		content := resp
		if i.FromURL {
			content = url
		}
		vals = append(vals, i.Rule.GetFirst(content))
	}
	c.result <- vals
}
