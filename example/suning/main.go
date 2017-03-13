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
	PageURLFmt = "http://zc.suning.com/project/browseList.htm?c=%s&t=&s=&keyWords=&pageNumber=%d"
)

var (
	Categories = map[string]string{
		"01": "科技",
		// "02": "设计",
		// "03": "公益",
		// "04": "农业",
		// "05": "文化",
		// "06": "娱乐",
		// "07": "其他",
	}

	StartPagesOfCategory = map[string]int{
		"01": 1,
		"02": 1,
		"03": 1,
		"04": 1,
		"05": 1,
		"06": 1,
		"07": 1,
	}
	EndPageOfCategory = map[string]int{
		"01": 20,
		"02": 9,
		"03": 2,
		"04": 16,
		"05": 4,
		"06": 5,
		"07": 3,
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
			Rule: ruler.NewCutStringRule(`<h1 class="item-detail-title">`, `</h1>`, nil),
		},
		&crawler.FieldItem{
			Name: "众筹状态",
			Rule: ruler.NewXPathNodeRule(`/html/body/div[5]/div/div[2]/div[4]/div[1]`, nil),
		},
		&crawler.FieldItem{
			Name: "目标金额",
			Rule: ruler.NewXPathNodeRule(
				`/html/body/div[5]/div/div[2]/div[4]/div[3]/strong[2]`,
				func(s string) string {
					s = strings.TrimPrefix(s, "¥")
					s = strings.TrimSpace(s)
					return strings.Replace(s, "\n", "", -1)
				}),
		},
		&crawler.FieldItem{
			Name: "实际筹资额",
			Rule: ruler.NewXPathNodeRule(`/html/body/div[5]/div/div[2]/div[4]/div[2]/span[3]`,
				func(s string) string {
					s = strings.TrimSpace(s)
					return strings.Replace(s, "\n", "", -1)
				}),
		},
		&crawler.FieldItem{
			Name: "项目支持人数",
			Rule: ruler.NewCutStringRule(
				`<div class="item-actor-num">已有<strong>`,
				`</strong>人支持该项目</div>`,
				func(s string) string {
					return strings.TrimSpace(s)
				}),
		},
		&crawler.FieldItem{
			Name: "项目回报种类",
			Rule: ruler.NewCutStringRule(
				`<div class="item-support-level">`,
				`<div class="item-support-risk">`,
				func(s string) string {
					return strconv.Itoa(strings.Count(s, `name="zc_detail_support_`))
				}),
		},

		&crawler.FieldItem{
			Name: "最低投资额",
			Rule: ruler.NewCutStringRule(
				`<div class="item-support-level">`,
				`<div class="item-support-risk">`,
				func(s string) string {
					rule := ruler.NewCutStringRule(
						`<strong class="price">`,
						`</strong>`,
						func(s string) string {
							s = strings.TrimSpace(s)
							return strings.Replace(s, "\n", "", -1)
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
			Rule: ruler.NewCutStringRule(
				`<div class="item-detail-info`,
				`<div class="item-topic`,
				func(s string) string {
					return strconv.Itoa(strings.Count(s, `<img`))
				}),
		},
		&crawler.FieldItem{
			Name: "关注数",
			Rule: ruler.NewCutStringRule(
				`<span name="zc_detail_guanzhu_01">关注 <i>`,
				`</i></span>`,
				func(s string) string {
					if s == "" {
						return "0"
					}

					return s
				},
			),
		},
		&crawler.FieldItem{
			Name: "话题数",
			Rule: ruler.NewXPathNodeRule(
				`/html/body/div[5]/div/div[1]/div/ul/li[2]/a/span`,
				func(s string) string {
					if s == "" {
						return "0"
					}

					return s
				},
			),
		},

		&crawler.FieldItem{
			Name: "是否制作视频",
			Rule: ruler.NewNooptRule(
				func(s string) string {
					if !strings.Contains(s, `title="Adobe Flash Player"`) {
						return "否"
					}
					return "是"
				}),
		},
		&crawler.FieldItem{
			Name: "预计回报发送时间",
			Rule: ruler.NewCutStringRule(`<p class="detail-time">预计发放时间：`, `</p>`, nil),
		},
		&crawler.FieldItem{
			Name: "最低投资支持人数",
			Rule: ruler.NewCutStringRule(
				`<div class="item-support-level">`,
				`<div class="item-support-risk">`,
				func(s string) string {
					rulePrice := ruler.NewCutStringRule(
						`<strong class="price">`,
						`</strong>`,
						func(s string) string {
							return strings.TrimSpace(s)
						},
					)
					prices := rulePrice.Get(s, false)

					rulePerson := ruler.NewCutStringRule(
						`无限额，`,
						`位支持者`,
						func(s string) string {
							return strings.TrimSpace(s)
						},
					)
					persons := rulePerson.Get(s, false)

					fmt.Println(">>>>>>>>>>>", prices, persons)
					count := len(prices)
					if count > len(persons) {
						count = len(prices)
					}
					minPrice, minId, found := 0.0, 0, false
					for i := 0; i < count; i++ {
						p := prices[i]
						pi, err := strconv.ParseFloat(p, 64)
						if err == nil && (!found || pi < minPrice) && pi != 1.0 {
							minId = i
							found = true
							minPrice = pi
						}
					}
					if minId >= len(persons) {
						return ""
					}
					return persons[minId]
				}),
		},
		&crawler.FieldItem{
			Name: "发起人所在地",
			Rule: ruler.NewCutStringRule(`"end_date": "`, `",`, nil),
		},
		&crawler.FieldItem{
			Name: "项目结束时间",
			Rule: ruler.NewXPathNodeRule(
				`/html/body/div[5]/div/div[2]/div[4]/div[5]`,
				func(s string) string {
					r := strings.NewReplacer("\n", "", " ", "")
					return strings.TrimSpace(r.Replace(s))
				}),
		},
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
				return fmt.Sprintf(PageURLFmt, codetmp, p)
			},
			startPage,
			endPage,
		)

		pageRule := ruler.NewRegexStringRule(
			`/project/detail.htm\?projectId=[0-9]+`,
			func(s string) string {
				return "http://zc.suning.com" + s
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
