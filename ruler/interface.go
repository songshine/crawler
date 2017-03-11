package ruler

type Interface interface {
	Get(content string, distinct bool) []string
	GetFirst(content string) string
}

type noopRule struct {
	trans transStringFunc
}

func NewNooptRule(trans transStringFunc) Interface {
	return &noopRule{
		trans: trans,
	}

}
func (r *noopRule) Get(content string, distinct bool) []string {
	return []string{r.trans.transString(content)}
}

func (r *noopRule) GetFirst(content string) string {
	return r.trans.transString(content)
}

type transStringFunc func(s string) string

func (t transStringFunc) transString(s string) string {
	if t != nil {
		return t(s)
	}

	return s
}

func (t transStringFunc) transStringSlice(ss []string) []string {
	if t != nil {
		result := make([]string, 0, len(ss))
		for _, s := range ss {
			result = append(result, t(s))
		}
		return result
	}

	return ss
}
