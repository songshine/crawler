package main

import (
	"fmt"
	"strings"

	"github.com/songshine/crawler"
	"github.com/songshine/crawler/phantom"
	"github.com/songshine/crawler/ruler"
)

const (
	StartCrawlerURL       = "http://z.jd.com/bigger/search.html"
	PaginationPostDataFmt = "status=&sort=&categoryId=%s&parentCategoryId=&sceneEnd=&productEnd=&keyword=&page=%d"
	ProjectRegexMatch     = `<a href="/project/details/[0-9]+.html"`
	ProjectURLFmt         = "http://z.jd.com/project/details/%s.html"
	UserCenterFmt         = "http://z.jd.com/funderCenter.action?flag=2&id=%s"
)

var (
	Categories = map[string]string{
		// "10": "科技", //3200
		//"36": "美食", //672
		// "37": "家电", //544
		//"12": "设计",  //1280
		//"11": "娱乐", //464 -
		//"38": "出版", //240
		"13": "公益", //304 -
		//"14": "其他", //1360 -
	}

	StartPagesOfCategory = map[string]int{
		"10": 95,
		"36": 31,
		"37": 1,
		"12": 58,
		"11": 1,
		"38": 4,
		"13": 11,
		"14": 1,
	}
	EndPageOfCategory = map[string]int{
		"10": 200,
		"36": 42,
		"37": 34,
		"12": 80,
		"11": 29,
		"38": 15,
		"13": 19,
		"14": 85,
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
		&crawler.FieldItem{
			Name:    "话题数",
			FromURL: true,
			Rule: ruler.NewEvaluationJSRule(
				`document.getElementById("topicBtn").childNodes[1].innerHTML`,
				`document.getElementById("topicBtn").childNodes[1].innerHTML !== "0"`,
				5000,
				func(s string) string {
					return strings.TrimSpace(s)
				},
			),
		},
		&crawler.FieldItem{
			Name:    "关注数",
			FromURL: true,
			Rule: ruler.NewEvaluationJSRule(
				`document.getElementById("focusCount").innerHTML`,
				`document.getElementById("focusCount").innerHTML !== "(0)"`,
				5000,
				func(s string) string {
					return strings.TrimSpace(s)
				},
			),
		},
	}
}

func main() {
	// gtk.Init(nil)
	// go func() {
	// 	runtime.LockOSThread()
	// 	gtk.Main()
	// }()

	getProjectNumberRuler := ruler.NewCutStringRule(
		`details/`,
		`.html`,
		func(s string) string {
			return fmt.Sprintf(ProjectURLFmt, s)
		},
	)

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
		pager := crawler.NewPostListPager(
			StartCrawlerURL,
			func(p int) string {
				return fmt.Sprintf(PaginationPostDataFmt, codetmp, p)
			},
			startPage,
			endPage,
		)

		pageRule := ruler.NewRegexStringRule(
			ProjectRegexMatch,
			func(s string) string {
				return getProjectNumberRuler.GetFirst(s)
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
		s := crawler.NewCSVDataStorage(fmt.Sprintf("%s_data_%d.csv", nametmp, startPage))
		s.Persist(dataCollector)
	}

	fmt.Println(">>>>>>>>>>>>>>>>>>>>>>>>>>>>Game over...")
	phantom.Exit()
}
