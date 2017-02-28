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
	var result []string
	si := strings.Index(content, r.start)
	if si == -1 {
		return result
	}
	si += len(r.start)
	ei := strings.Index(content, r.end)

	if ei == -1 {
		return result
	}

	if ei <= si || ei >= len(content) {
		return result
	}

	match := content[si:ei]
	if match == "" {
		return result
	}
	result = append(result, r.trans.transString(match))
	ei += len(r.end)
	if ei >= len(content) {
		return result
	}
	subs := r.Get(content[ei:], distinct)
	if !distinct {
		result = append(result, subs...)
		return result
	}
	dupCheck := make(map[string]struct{})
	for _, s := range subs {
		if _, ok := dupCheck[s]; ok {
			continue
		}
		dupCheck[s] = struct{}{}
		result = append(result, s)
	}
	return result
}

func (r *cutStringRule) GetFirst(content string) string {
	s := strings.Index(content, r.start)
	if s == -1 {
		return ""
	}
	s += len(r.start)
	e := strings.Index(content, r.end)
	if e > s && e < len(content) {
		return r.trans.transString(content[s:e])
	}
	return ""
}
