package ruler

import "strings"

func NewCutStringRule(start, end string, transFunc transStringFunc) Interface {
	r := &cutStringRule{
		start: start,
		end:   end,
		trans: transFunc,
	}
	return r
}

// implement ExtractStringRuler
type cutStringRule struct {
	start, end string
	trans      transStringFunc
}

func (r *cutStringRule) Get(content string, distinct bool) []string {
	var ts []string
	for {
		s, i := r.getNext(content)
		if i == -1 {
			break
		}
		ts = append(ts, s)
		content = content[i:]

	}

	if !distinct {
		return ts
	}

	var result []string
	dupCheck := make(map[string]struct{})
	for _, s := range ts {
		if _, ok := dupCheck[s]; ok {
			continue
		}
		dupCheck[s] = struct{}{}
		result = append(result, s)
	}
	return result
}

func (r *cutStringRule) GetFirst(content string) string {
	s, _ := r.getNext(content)
	return s
}

func (r *cutStringRule) getNext(content string) (string, int) {
	s := strings.Index(content, r.start)
	if s == -1 {
		return r.trans.transString(""), -1
	}
	s += len(r.start)
	if s >= len(content) {
		return r.trans.transString(""), -1
	}

	e := strings.Index(content[s:], r.end)

	if e == -1 {
		return r.trans.transString(""), -1
	}
	return r.trans.transString(content[s : s+e]), s + e
}
