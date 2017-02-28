package main

import (
	"fmt"
	"strings"
	"strconv"
	"github.com/songshine/crawler"
	"github.com/songshine/crawler/ruler"

)

const (
	StartCrawlerURL       = "https://z.jd.com/bigger/search.html"
	PaginationPostDataFmt = "status=&sort=&categoryId=%s&parentCategoryId=&sceneEnd=&productEnd=&keyword=&page=%d"
	ProjectRegexMatch     = `<a href="/project/details/[0-9]+.html"`
	ProjectURLFmt         = "https://z.jd.com/project/details/%s.html"
)

var (
	Categories = map[string]string{
		"10": "科技",
	}

	TotalPagesOfCategory = map[string]int{
		"10": 10,
	}

)

func buildFieldRules() []*crawler.FieldItem{
	return []*crawler.FieldItem {
		&crawler.FieldItem{
			Name: "项目编号",
			FromURL: true,
			Rule: ruler.NewRegexStringRule("[0-9]+", nil),
		},
		&crawler.FieldItem{
			Name: "项目名称",
			Rule: ruler.NewCutStringRule(`<p class="p-title">`, `</p>`, nil),
		},
		&crawler.FieldItem{
			Name:"发起人",
			FromURL: true,
			Rule: ruler.NewRegexStringRule(
				"[0-9]+", 
				func(s string) string {
					newURL := fmt.Sprintf(`https://z.jd.com/funderCenter.action?flag=2&id=%s`, s)
					rule := ruler.NewXPathNodeRule(
						`//*[@id="mainframe"]/div[2]/div[1]/div[1]/div[1]/div/div[1]/p`,
						func(s string)string{
							return strings.TrimSpace(s)
						},
					)
					return crawler.GetFromNextPage(newURL, rule)
				},
			),
		},
		&crawler.FieldItem{
			Name:"发起人发起的项目",
			FromURL: true,
			Rule: ruler.NewRegexStringRule(
				"[0-9]+", 
				func(s string) string {
					newURL := fmt.Sprintf(`https://z.jd.com/funderCenter.action?flag=2&id=%s`, s)
					rule := ruler.NewXPathNodeRule(
						`//*[@id="mainframe"]/div[2]/div[1]/div[1]/div[2]/a[2]/i`,
						func(s string)string{
							return strings.TrimSpace(s)
						},
					)
					return crawler.GetFromNextPage(newURL, rule)
				},
			),
		},
		&crawler.FieldItem{
			Name: "图片数量",
			Rule: ruler.NewCutStringRule(
				`<!--图片部分-->`, 
				`<!--图片部分end-->`, 
				func(s string) string {
					return strconv.Itoa(strings.Count(s, `<img alt`))
				},
			),
		},
		
	}
}

func main() {
	getProjectNumberRuler := ruler.NewCutStringRule(
		`details/`,
		`.html`,
		func(s string) string {
			return fmt.Sprintf(ProjectURLFmt, s)
		},
	)
	for code := range Categories {
		totalPage := TotalPagesOfCategory[code]
		codetmp := code
		pager := crawler.NewPostListPager(
			StartCrawlerURL,
			func(p int) string {
				return fmt.Sprintf(PaginationPostDataFmt, codetmp, p)
			},
			1,
			totalPage,
		)

		ruler := ruler.NewRegexStringRule(
			ProjectRegexMatch,
			func(s string) string {
				return getProjectNumberRuler.GetFirst(s)
			},
		)
		dataCollector := crawler.NewDataCollector(
			crawler.NewURLCollector(pager, ruler),
			buildFieldRules()...
		)
		s := crawler.NewCSVDataStorage("test.csv")
		s.Persist(dataCollector)
	}

}
