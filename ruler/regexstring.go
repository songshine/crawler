package ruler

import (
	"fmt"
	"regexp"
)

func NewRegexStringRule(regex string, transFunc transStringFunc) Interface {
	r := &regexStringMatchRule{
		match: regex,
		trans: transFunc,
	}
	return r
}

// implement ExtractStringRuler
type regexStringMatchRule struct {
	match string
	trans transStringFunc
}

func (r *regexStringMatchRule) Get(content string, distinct bool) []string {
	rex := regexp.MustCompile(r.match)
	matches := rex.FindAllString(content, -1)
	fmt.Println(content, r.match, matches)
	if !distinct {
		return r.trans.transStringSlice(matches)
	}

	var (
		result   = make([]string, 0, len(matches))
		dupCheck = make(map[string]struct{})
	)
	for _, m := range matches {
		if _, ok := dupCheck[m]; ok {
			continue
		}
		dupCheck[m] = struct{}{}
		result = append(result, m)
	}

	return r.trans.transStringSlice(result)
}

func (r *regexStringMatchRule) GetFirst(content string) string {
	rex := regexp.MustCompile(r.match)
	match := rex.FindString(content)
	return r.trans.transString(match)
}
