package main

import (
	"fmt"
	"log"
	"strings"

	"strconv"

	"github.com/songshine/crawler"
	"github.com/songshine/crawler/ruler"
)

const (
	PageURLFmt = `http://www.dreamore.com/projects?status=%d&type=%s&p=%d`
)

var (
	Categories = map[string]string{
		"20": "设计",
		"21": "科技",
		"22": "影像",
		"23": "音乐",
		"24": "人文",
		"25": "出版",
		"26": "活动",
		"27": "其他",
	}

	StartPagesOfCategory = map[string]int{
		"20": 1,
		"21": 1,
		"22": 1,
		"23": 1,
		"24": 1,
		"25": 1,
		"26": 1,
		"27": 1,
	}
	EndPageOfCategory = map[string]int{
		"20": 6,
		"21": 3,
		"22": 5,
		"23": 2,
		"24": 6,
		"25": 3,
		"26": 6,
		"27": 6,
	}

	Status = map[int]string{
		3: "众筹中",
		4: "已成功",
		5: "已结束",
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
			Rule: ruler.NewXPathNodeRule(`/html/body/section/div[1]/div/div[1]/h5/span`, nil),
		},
		&crawler.FieldItem{
			Name: "目标金额",
			Rule: ruler.NewXPathNodeRule(`/html/body/section/div[1]/div/div[2]/div[2]/div[5]/span[1]`, func(s string) string { return strings.TrimPrefix(s, "¥") }),
		},
		&crawler.FieldItem{
			Name: "实际筹资额",
			Rule: ruler.NewXPathNodeRule(`//*[@id="refresh_money"]`, func(s string) string { return strings.TrimPrefix(s, "¥") }),
		}, ///html/body/section/div[2]/div/div[1]/div[1]/a[4]/span
		&crawler.FieldItem{
			Name: "项目支持人数",
			Rule: ruler.NewXPathNodeRule(`html/body/section/div[2]/div/div[1]/div[1]/a[4]/span`, func(s string) string {
				ss := strings.TrimPrefix(s, "支持人次（")
				return strings.TrimSuffix(ss, "）")
			}),
		}, //
		&crawler.FieldItem{
			Name: "项目回报总类",
			Rule: ruler.NewCutStringRule(`<div class="pindex_return_top">`, `</div>`, func(s string) string {
				return strconv.Itoa(strings.Count(s, "<li name="))
			}),
		},
		&crawler.FieldItem{
			Name: "最低投资额",
			Rule: ruler.NewCutStringRule(`<div class="pindex_return_top">`, `</div>`, func(s string) string {
				rule := ruler.NewCutStringRule(
					`¥`,
					`</li>`,
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
			Rule: ruler.NewCutStringRule(`<div class="pindex_con_content">`, `div>`, func(s string) string { return strconv.Itoa(strings.Count(s, `<img `)) }),
		},
		&crawler.FieldItem{
			Name: "项目进展数",
			Rule: ruler.NewCutStringRule(`项目进展（`, `）`, nil),
		},
		&crawler.FieldItem{
			Name: "评论数",
			Rule: ruler.NewCutStringRule(`评论（`, `）`, nil),
		},
		&crawler.FieldItem{
			Name: "是否制作视频",
			Rule: ruler.NewNooptRule(func(s string) string {
				if strings.Contains(s, `id="youkuplayer"`) {
					return "是"
				}
				return "否"
			}),
		},
		&crawler.FieldItem{
			Name: "预计回报发送时间",
			Rule: ruler.NewCutStringRule(`预计成功之后 `, ` 天内寄送`, func(s string) string { return s + "天内" }),
		},
		&crawler.FieldItem{
			Name: "发起人支持的项目数",
			Rule: ruler.NewRegexStringRule(
				`http://www.dreamore.com/space/[0-9]+`,
				func(s string) string {
					rule := ruler.NewNooptRule(func(s string) string {
						return strconv.Itoa(strings.Count(s, `data-id="`))
					})

					return crawler.GetFromNextPage(s+"/support", rule)
				}),
		},
		&crawler.FieldItem{
			Name: "发起人历史发起的项目次数",
			Rule: ruler.NewRegexStringRule(
				`http://www.dreamore.com/space/[0-9]+`,
				func(s string) string {
					rule := ruler.NewNooptRule(func(s string) string {
						return strconv.Itoa(strings.Count(s, `data-id="`))
					})

					return crawler.GetFromNextPage(s+"/publish", rule)
				}),
		},

		&crawler.FieldItem{
			Name: "发起人发起的活动数",
			Rule: ruler.NewRegexStringRule(
				`http://www.dreamore.com/space/[0-9]+`,
				func(s string) string {
					rule := ruler.NewNooptRule(func(s string) string {
						return strconv.Itoa(strings.Count(s, `data-id="`))
					})

					return crawler.GetFromNextPage(s+"/event", rule)
				}),
		},

		&crawler.FieldItem{
			Name: "发起人参与的活动数",
			Rule: ruler.NewRegexStringRule(
				`http://www.dreamore.com/space/[0-9]+`,
				func(s string) string {
					rule := ruler.NewNooptRule(func(s string) string {
						return strconv.Itoa(strings.Count(s, `data-id="`))
					})

					return crawler.GetFromNextPage(s+"/part", rule)
				}),
		},
		&crawler.FieldItem{
			Name: "发起人所在地",
			Rule: ruler.NewXPathNodeRule(`/html/body/section/div[2]/div/div[2]/div[1]/div[2]/p[1]/span/span[2]`, nil),
		},
		&crawler.FieldItem{
			Name: "项目结束时间",
			Rule: ruler.NewXPathNodeRule(
				`/html/body/section/div[2]/div/div[2]/div[1]/div[2]/p[3]/span/span[2]`,
				func(s string) string {
					return strings.TrimSpace(s)
				},
			),
		},
		&crawler.FieldItem{
			Name: "发起人积分",
			Rule: ruler.NewRegexStringRule(
				`http://www.dreamore.com/space/[0-9]+`,
				func(s string) string {
					rule := ruler.NewXPathNodeRule(`/html/body/section/div/div[1]/div[2]/p[3]/span/span[2]`, nil)
					return crawler.GetFromNextPage(s, rule)
				}),
		},
		&crawler.FieldItem{
			Name: "最低投资支持人数",
			Rule: ruler.NewCutStringRule(`"items":`, `]`,
				func(s string) string {
					rulePrice := ruler.NewCutStringRule(
						`¥`,
						`</li>`,
						func(s string) string {
							return strings.TrimSpace(s)
						},
					)
					prices := rulePrice.Get(s, false)

					rulePerson := ruler.NewCutStringRule(
						`class="money_12 y">`,
						`</span>`,
						func(s string) string {
							return strings.TrimSpace(s)
						},
					)
					persons := rulePerson.Get(s, false)

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

		for sc, sd := range Status {
			pager := crawler.NewGetListPager(
				func(p int) string {
					return fmt.Sprintf(PageURLFmt, sc, codetmp, p)
				},
				startPage,
				endPage,
			)

			pageRule := ruler.NewRegexStringRule(
				`http://www.dreamore.com/projects/[0-9]+.html`,
				nil,
			)
			fieldItems := []*crawler.FieldItem{
				&crawler.FieldItem{
					Name: "项目类型",
					Rule: ruler.NewConstStringRule(nametmp, nil),
				},
				&crawler.FieldItem{
					Name: "众筹状态",
					Rule: ruler.NewConstStringRule(sd, nil),
				},
			}
			fieldItems = append(fieldItems, buildFieldRules()...)
			dataCollector := crawler.NewDataCollector(
				crawler.NewURLCollector(pager, pageRule),
				fieldItems...,
			)
			s := crawler.NewCSVDataStorage(fmt.Sprintf("%s_%s_data.csv", nametmp, sd))
			s.Persist(dataCollector)
		}

	}

	log.Println(">>> Completed successfully!!")
}
