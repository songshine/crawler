package ruler

func NewConstStringRule(cons string, transFunc transStringFunc) Interface {
	r := &constStringRule{
		cst:   cons,
		trans: transFunc,
	}
	return r
}

// implement ExtractStringRuler
type constStringRule struct {
	cst   string
	trans transStringFunc
}

func (r *constStringRule) Get(content string, distinct bool) []string {
	return r.trans.transStringSlice([]string{r.cst})
}

func (r *constStringRule) GetFirst(content string) string {
	return r.trans.transString(r.cst)
}
