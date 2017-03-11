package main

import (
	"fmt"
	"log"
	"strconv"

	"strings"

	"github.com/songshine/crawler"
	"github.com/songshine/crawler/ruler"
)

const (
	PageURLFmt = "http://izhongchou.taobao.com/dream/ajax/getProjectList.htm?page=%d&pageSize=20&projectType=%s&type=6&status=&sort=1"
)

var (
	Categories = map[string]string{
		"121288001": "科技",
		// "123330001": "农业",
		// "122018001": "动漫",
		// "121292001": "设计",
		// "121280001": "公益",
		// "121284001": "娱乐",
		// "121278001": "影音",
		// "121274002": "书籍",
		// "122020001": "游戏",
		// "123332001": "其他",
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
		&crawler.FieldItem{
			Name: "项目名称",
			Rule: ruler.NewCutStringRule(`"name":"`, `"`, func(s string) string { return crawler.Unicode2UTF8(s) }),
		},
		&crawler.FieldItem{
			Name: "众筹状态",
			Rule: ruler.NewCutStringRule(`"status":"`, `"`, func(s string) string { return crawler.Unicode2UTF8(s) }),
		},
		&crawler.FieldItem{
			Name: "目标金额",
			Rule: ruler.NewCutStringRule(`"target_money": "`, `"`, func(s string) string { return crawler.Unicode2UTF8(s) }),
		},
		&crawler.FieldItem{
			Name: "实际筹资额",
			Rule: ruler.NewCutStringRule(`"curr_money": "`, `"`, func(s string) string { return crawler.Unicode2UTF8(s) }),
		},
		&crawler.FieldItem{
			Name: "项目支持人数",
			Rule: ruler.NewCutStringRule(`"support_person": "`, `"`, func(s string) string { return crawler.Unicode2UTF8(s) }),
		},
		&crawler.FieldItem{
			Name: "项目回报种类",
			Rule: ruler.NewCutStringRule(`"items":`, `]`, func(s string) string { return strconv.Itoa(strings.Count(s, `"item_id"`)) }),
		},

		&crawler.FieldItem{
			Name: "最低投资额",
			Rule: ruler.NewCutStringRule(`"items":`, `]`,
				func(s string) string {
					rule := ruler.NewCutStringRule(
						`"price": "`,
						`"`,
						func(s string) string {
							return strings.TrimSpace(s)
						},
					)
					prices := rule.Get(s, true)
					minPrice, found := 0.0, false
					for _, p := range prices {
						pi, err := strconv.ParseFloat(p, 64)
						if err == nil && (!found || pi < minPrice) && pi != 1.0 {
							found = true
							minPrice = pi
						}
					}
					return strconv.FormatFloat(minPrice, 'f', 2, 64)
				}),
		},
		&crawler.FieldItem{
			Name: "项目图片数量",
			Rule: ruler.NewNooptRule(func(s string) string { return strconv.Itoa(strings.Count(s, `IMG src=`)) }),
		},
		&crawler.FieldItem{
			Name:    "项目进展数",
			FromURL: true,
			Rule: ruler.NewRegexStringRule("[0-9]+", func(s string) string {
				url := fmt.Sprintf(`https://izhongchou.taobao.com/dream/ajax/get_project_feeds.htm?project_id=%s&_tb_token_=xDEcRdmNPq`, "20056473")
				res, err := crawler.Get(url)
				if err != nil {

					return "err"
				}
				fmt.Println(url, res)
				return strconv.Itoa(strings.Count(res, `"feed_id":`))
			}),
		},
		// &crawler.FieldItem{
		// 	Name: "喜欢数",
		// 	Rule: ruler.NewCutStringRule(`"items":`, `]`, func(s string) string { return strconv.Itoa(strings.Count(s, `"item_id"`)) }),
		// },
		// &crawler.FieldItem{
		// 	Name: "项目结束时间",
		// 	Rule: ruler.NewCutStringRule(`"items":`, `]`, func(s string) string { return strconv.Itoa(strings.Count(s, `"item_id"`)) }),
		// },
		// &crawler.FieldItem{
		// 	Name: "是否制作视频",
		// 	Rule: ruler.NewCutStringRule(`"items":`, `]`, func(s string) string { return strconv.Itoa(strings.Count(s, `"item_id"`)) }),
		// },
		// &crawler.FieldItem{
		// 	Name: "最低投资支持人数",
		// 	Rule: ruler.NewCutStringRule(`"items":`, `]`, func(s string) string { return strconv.Itoa(strings.Count(s, `"item_id"`)) }),
		// },
		// &crawler.FieldItem{
		// 	Name: "发起人所在地",
		// 	Rule: ruler.NewCutStringRule(`"items":`, `]`, func(s string) string { return strconv.Itoa(strings.Count(s, `"item_id"`)) }),
		// },
		// &crawler.FieldItem{
		// 	Name: "发起人积分",
		// 	Rule: ruler.NewCutStringRule(`"items":`, `]`, func(s string) string { return strconv.Itoa(strings.Count(s, `"item_id"`)) }),
		// },
		// &crawler.FieldItem{
		// 	Name: "是否有专利证书",
		// 	Rule: ruler.NewCutStringRule(`"items":`, `]`, func(s string) string { return strconv.Itoa(strings.Count(s, `"item_id"`)) }),
		// },
		// &crawler.FieldItem{
		// 	Name: "个人还是团队",
		// 	Rule: ruler.NewCutStringRule(`"items":`, `]`, func(s string) string { return strconv.Itoa(strings.Count(s, `"item_id"`)) }),
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
			`//izhongchou.taobao.com/dreamdetail.htm\?id=[0-9]+`,
			func(s string) string {
				url := "http:" + s
				crawler.Get(url)
				rule := ruler.NewRegexStringRule("id=[0-9]+", nil)
				return `https://izhongchou.taobao.com/dream/ajax/getProjectForDetail.htm?` + rule.GetFirst(s)
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
}
