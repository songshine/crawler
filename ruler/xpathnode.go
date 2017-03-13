package ruler

import (
	"bytes"
	"log"
	"strings"

	"github.com/go-xmlpath/xmlpath"
	"golang.org/x/net/html"
)

func NewXPathNodeRule(xpath string, transFunc transStringFunc) Interface {
	r := &xPathNodeRule{
		xpath: xpath,
		trans: transFunc,
	}
	return r
}

type xPathNodeRule struct {
	xpath string
	trans transStringFunc
}

func (r *xPathNodeRule) Get(content string, distinct bool) []string {
	val := r.GetFirst(content)
	if val != "" {
		return []string{val}
	}
	return nil
}

func (r *xPathNodeRule) GetFirst(content string) string {
	reader := strings.NewReader(content)
	root, err := html.Parse(reader)

	if err != nil {
		log.Printf("Invalid html: %s, error: %v\n", content, err)
		return r.trans.transString("")
	}

	var b bytes.Buffer
	html.Render(&b, root)

	fixedHTML := b.String()

	reader = strings.NewReader(fixedHTML)
	xmlroot, xmlerr := xmlpath.ParseHTML(reader)

	if xmlerr != nil {
		log.Printf("Invalid html: %s, error: %v\n", content, xmlerr)
		return r.trans.transString("")
	}

	path := xmlpath.MustCompile(r.xpath)
	if value, ok := path.String(xmlroot); ok {
		return r.trans.transString(value)
	}
	return r.trans.transString("")
}
