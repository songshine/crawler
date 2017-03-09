package crawler

type fmtPageFunc func(int) string

type ListPager interface {
	Next() (string, bool)
}

type defaultListPager struct {
	respChan chan string
}

func newDefaultListPager() *defaultListPager {
	return &defaultListPager{
		respChan: make(chan string, 100),
	}
}

func (p *defaultListPager) Next() (resp string, done bool) {
	resp, ok := <-p.respChan
	return resp, !ok
}

// implement ListPager
type listPagerPost struct {
	*defaultListPager
	url              string
	pageFunc         fmtPageFunc
	fromPage, toPage int
}

func NewPostListPager(url string, pageFunc fmtPageFunc, from, to int) ListPager {
	p := &listPagerPost{
		url:              url,
		fromPage:         from,
		toPage:           to,
		pageFunc:         pageFunc,
		defaultListPager: newDefaultListPager(),
	}
	go func() {
		for s := p.fromPage; s <= p.toPage; s++ {
			body := p.pageFunc(s)
			resp, err := PostString(p.url, body)
			if err != nil {
				continue
			}
			p.respChan <- resp

		}
		close(p.respChan)
	}()

	return p
}

type listPagerGet struct {
	*defaultListPager
	pageURLFunc      fmtPageFunc
	fromPage, toPage int
}

func NewGetListPager(pageURLFunc fmtPageFunc, from, to int) ListPager {
	p := &listPagerGet{
		fromPage:         from,
		toPage:           to,
		pageURLFunc:      pageURLFunc,
		defaultListPager: newDefaultListPager(),
	}

	go func() {
		for s := p.fromPage; s <= p.toPage; s++ {
			url := p.pageURLFunc(s)
			resp, err := Get(url)
			if err != nil {
				continue
			}
			p.respChan <- resp

		}
		close(p.respChan)
	}()

	return p
}
