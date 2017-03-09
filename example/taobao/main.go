package main

import (
	"fmt"
	"log"

	"github.com/songshine/crawler"
	"github.com/songshine/crawler/phantom"
	"github.com/songshine/crawler/ruler"
)

const (
	PageURLFmt = "http://izhongchou.taobao.com/dream/ajax/getProjectList.htm?page=%d&pageSize=20&projectType=%s&type=6&status=&sort=1"
)

var (
	Categories = map[string]string{
		"121288001": "科技",
		"123330001": "农业",
		"122018001": "动漫",
		"121292001": "设计",
		"121280001": "公益",
		"121284001": "娱乐",
		"121278001": "影音",
		"121274002": "书籍",
		"122020001": "游戏",
		"123332001": "其他",
	}

	StartPagesOfCategory = map[string]int{
		"121288001": 1,
		"123330001": 1,
		"122018001": 1,
		"121292001": 1,
		"121280001": 1,
		"121284001": 1,
		"121278001": 1,
		"121274002": 1,
		"122020001": 1,
		"123332001": 1,
	}
	EndPageOfCategory = map[string]int{
		"121288001": 101,
		"123330001": 115,
		"122018001": 23,
		"121292001": 67,
		"121280001": 20,
		"121284001": 10,
		"121278001": 11,
		"121274002": 3,
		"122020001": 2,
		"123332001": 9,
	}
)

func buildFieldRules() []*crawler.FieldItem {
	return []*crawler.FieldItem{
		&crawler.FieldItem{
			Name:    "项目编号",
			FromURL: true,
			Rule:    ruler.NewRegexStringRule("[0-9]+", nil),
		},
		// &crawler.FieldItem{
		// 	Name: "项目名称",
		// 	Rule: ruler.NewCutStringRule(`<p class="p-title">`, `</p>`, nil),
		// },
		// &crawler.FieldItem{
		// 	Name: "项目回报总类",
		// 	Rule: ruler.NewCutStringRule(
		// 		`<!-- 档位 -->`,
		// 		`<!--price-box无私奉献-->`,
		// 		func(s string) string {
		// 			return strconv.Itoa(strings.Count(s, `<!--price-box-->`))
		// 		},
		// 	),
		// },
		// &crawler.FieldItem{
		// 	Name: "最低投资额",
		// 	Rule: ruler.NewCutStringRule(
		// 		`<!-- 档位 -->`,
		// 		`<!--price-box无私奉献-->`,
		// 		func(s string) string {
		// 			rule := ruler.NewCutStringRule(
		// 				`<!--price-box-->`,
		// 				`<!--price-box end-->`,
		// 				func(s string) string {
		// 					if strings.Contains(s, "抽奖档") {
		// 						return ""
		// 					}
		// 					rule := ruler.NewCutStringRule(
		// 						`￥<span>`,
		// 						`</span>`,
		// 						nil,
		// 					)

		// 					return strings.TrimSpace(rule.GetFirst(s))
		// 				},
		// 			)

		// 			prices := rule.Get(s, false)
		// 			minPrice, found := 0, false
		// 			for _, p := range prices {
		// 				pi, err := strconv.Atoi(p)
		// 				if err == nil && (!found || pi < minPrice) {
		// 					found = true
		// 					minPrice = pi
		// 				}
		// 			}
		// 			return strconv.Itoa(minPrice)
		// 		},
		// 	),
		// },
		// &crawler.FieldItem{
		// 	Name: "项目图片数量",
		// 	Rule: ruler.NewCutStringRule(
		// 		`<!--图片部分-->`,
		// 		`<!--图片部分end-->`,
		// 		func(s string) string {
		// 			return strconv.Itoa(strings.Count(s, `<img alt`))
		// 		},
		// 	),
		// },
		// &crawler.FieldItem{
		// 	Name:    "发起人支持的项目数",
		// 	FromURL: true,
		// 	Rule: ruler.NewRegexStringRule(
		// 		"[0-9]+",
		// 		func(s string) string {
		// 			newURL := fmt.Sprintf(UserCenterFmt, s)
		// 			rule := ruler.NewXPathNodeRule(
		// 				`//*[@id="mainframe"]/div[2]/div[1]/div[1]/div[2]/a[1]/i`,
		// 				func(s string) string {
		// 					return strings.TrimSpace(s)
		// 				},
		// 			)
		// 			return crawler.GetFromNextPage(newURL, rule)
		// 		},
		// 	),
		// },
		// &crawler.FieldItem{
		// 	Name:    "发起人历史发起的项目",
		// 	FromURL: true,
		// 	Rule: ruler.NewRegexStringRule(
		// 		"[0-9]+",
		// 		func(s string) string {
		// 			newURL := fmt.Sprintf(UserCenterFmt, s)
		// 			rule := ruler.NewXPathNodeRule(
		// 				`//*[@id="mainframe"]/div[2]/div[1]/div[1]/div[2]/a[2]/i`,
		// 				func(s string) string {
		// 					return strings.TrimSpace(s)
		// 				},
		// 			)
		// 			return crawler.GetFromNextPage(newURL, rule)
		// 		},
		// 	),
		// },
		// &crawler.FieldItem{
		// 	Name:    "话题数",
		// 	FromURL: true,
		// 	Rule: ruler.NewEvaluationJSRule(
		// 		`document.getElementById("topicBtn").childNodes[1].innerHTML`,
		// 		`document.getElementById("topicBtn").childNodes[1].innerHTML !== "0"`,
		// 		5000,
		// 		func(s string) string {
		// 			return strings.TrimSpace(s)
		// 		},
		// 	),
		// },
		// &crawler.FieldItem{
		// 	Name:    "关注数",
		// 	FromURL: true,
		// 	Rule: ruler.NewEvaluationJSRule(
		// 		`document.getElementById("focusCount").innerHTML`,
		// 		`document.getElementById("focusCount").innerHTML !== "(0)"`,
		// 		5000,
		// 		func(s string) string {
		// 			return strings.TrimSpace(s)
		// 		},
		// 	),
		// },
	}
}

func main() {
	for code, name := range Categories {
		startPage := StartPagesOfCategory[code]
		if startPage == 0 {
			continue
		}
		endPage := EndPageOfCategory[code]
		if endPage < startPage {
			continue
		}

		codetmp := code
		nametmp := name
		pager := crawler.NewGetListPager(
			func(p int) string {
				return fmt.Sprintf(PageURLFmt, p, codetmp)
			},
			startPage,
			endPage,
		)

		pageRule := ruler.NewRegexStringRule(
			`//izhongchou.taobao.com/dreamdetail.htm?id=[0-9]+`,
			func(s string) string {
				return "http:" + s
			},
		)
		fieldItems := []*crawler.FieldItem{
			&crawler.FieldItem{
				Name: "项目类型",
				Rule: ruler.NewConstStringRule(nametmp, nil),
			},
		}
		fieldItems = append(fieldItems, buildFieldRules()...)
		dataCollector := crawler.NewDataCollector(
			crawler.NewURLCollector(pager, pageRule),
			fieldItems...,
		)
		s := crawler.NewCSVDataStorage(fmt.Sprintf("%s_data.csv", nametmp))
		s.Persist(dataCollector)
	}

	log.Println(">>> Completed successfully!!")
	phantom.Exit()
}
