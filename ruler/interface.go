package ruler

type Interface interface {
	Get(content string, distinct bool) []string
	GetFirst(content string) string
}

type NoopRule struct {
	Trans transStringFunc
}

func (r *NoopRule) Get(content string, distinct bool) []string {
	return []string{r.Trans.transString(content)}
}

func (r *NoopRule) GetFirst(content string) string {
	return r.Trans.transString(content)
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
