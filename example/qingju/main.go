package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"time"

	"github.com/songshine/crawler"
	"github.com/songshine/crawler/ruler"
)

const (
	PageURLFmt = `http://www.qingju.com/discover?category_id=%s&page=%d`
)

var (
	Categories = map[string]string{
		// "1": "出版",
		// "3": "音乐",
		// "4": "影视",
		// "2": "设计",
		// "5": "游戏",
		// "6": "科技",
		// "8":  "公益",
		// "9":  "活动", // 142
		"11": "周边",
		// "12": "动漫",
		// "7": "其他",
	}

	StartPagesOfCategory = map[string]int{
		"1":  1,
		"3":  1,
		"4":  1,
		"2":  1,
		"5":  1,
		"6":  1,
		"8":  1,
		"9":  1,
		"11": 1,
		"12": 1,
		"7":  1,
	}
	EndPageOfCategory = map[string]int{
		"1":  2,
		"3":  1,
		"4":  2,
		"2":  4,
		"5":  1,
		"6":  3,
		"8":  1,
		"9":  5,
		"11": 1,
		"12": 1,
		"7":  4,
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
			Rule: ruler.NewCutStringRule(
				`<header class="text-center"><h2 class="bold">`,
				`</h2><h5>`,
				nil),
		},
		&crawler.FieldItem{
			Name: "众筹状态",
			Rule: ruler.NewNooptRule(
				func(s string) string {

					var (
						finishTime time.Time
						per        int
					)
					ruleTi := ruler.NewCutStringRule(
						`<p>若不能在 `,
						` 前筹集到`,
						nil,
					)
					finishTime, _ = time.Parse("2006-01-02 15:04", ruleTi.GetFirst(s))
					rulePer := ruler.NewCutStringRule(`筹资目标<span class="pull-right"> `, `%完成度</span></p>`, nil)
					if perStr := rulePer.GetFirst(s); perStr != "" {
						perStr = strings.TrimSuffix(perStr, " %完成度")
						if perStr != "" {
							per, _ = strconv.Atoi(perStr)
						}
					}

					now := time.Now()

					if finishTime.Unix() < now.Unix() && per >= 100 {
						return "筹集成功"
					} else if finishTime.Unix() < now.Unix() && per < 100 {
						return "筹集结束"
					} else {
						return "筹集中"
					}
				},
			),
		},
		&crawler.FieldItem{
			Name: "目标金额",
			Rule: ruler.NewCutStringRule(`<p>￥`, `筹资目标<span`, func(s string) string { return strings.Replace(s, ",", "", -1) }),
		},
		&crawler.FieldItem{
			Name: "实际筹资额",
			Rule: ruler.NewCutStringRule(
				`目前累计金额</p><p class="text-xlg bold">￥`, `</p>`,
				func(s string) string {
					return strings.Replace(s, ",", "", -1)
				}),
		},
		&crawler.FieldItem{
			Name: "项目支持人数",
			Rule: ruler.NewCutStringRule(
				`<div class="pull-right"><span class="project-stats-value">`,
				`</span>`,
				func(s string) string {
					return strings.Replace(s, ",", "", -1)
				}),
		},
		&crawler.FieldItem{
			Name: "项目回报总类",
			Rule: ruler.NewNooptRule(func(s string) string {
				return strconv.Itoa(strings.Count(s, `<div class="reward section">`))
			}),
		},
		&crawler.FieldItem{
			Name: "最低投资额",
			Rule: ruler.NewNooptRule(func(s string) string {
				rule := ruler.NewCutStringRule(
					`<p class="h3plus text-primary">支持`,
					`元或更多</p>`,
					func(s string) string {
						s = strings.TrimSpace(s)
						return strings.Replace(s, ",", "", -1)
					},
				)
				prices := rule.Get(s, true)
				minPrice, found := 0.0, false
				for _, p := range prices {
					pi, err := strconv.ParseFloat(p, 64)
					if err == nil && (!found || pi < minPrice) && pi > 1.0 {
						found = true
						minPrice = pi
					}
				}
				return strconv.FormatFloat(minPrice, 'f', 2, 64)

			}),
		},

		&crawler.FieldItem{
			Name: "项目图片数量",
			Rule: ruler.NewNooptRule(func(s string) string { return strconv.Itoa(strings.Count(s, `<img src`)) }),
		},
		&crawler.FieldItem{
			Name: "梦想动态",
			Rule: ruler.NewCutStringRule(
				`梦想动态<span class="badge">`,
				`</span></a>`,
				func(s string) string {
					if s == "" {
						return "0"
					}
					return strings.Replace(s, ",", "", -1)
				}),
		},
		&crawler.FieldItem{
			Name: "是否制作视频",
			Rule: ruler.NewNooptRule(func(s string) string {
				if strings.Contains(s, `main-video`) {
					return "是"
				}
				return "否"
			}),
		},

		&crawler.FieldItem{
			Name: "预计回报发送时间",
			Rule: ruler.NewCutStringRule(
				`包邮（大陆地区)&nbsp;&nbsp;&nbsp;`,
				`</div>`,
				func(s string) string {
					if s != "" {
						return strings.TrimSpace(s)
					}
					return "不确定"
				}),
		}, //<ul class="v-layout"><li>发起过1个项目，0次支持</li>
		&crawler.FieldItem{
			Name: "发起人支持的项目数",
			Rule: ruler.NewCutStringRule(
				`<ul class="v-layout"><li>发起过`,
				`</li>`,
				func(s string) string {
					rule := ruler.NewCutStringRule(
						`个项目，`,
						`次支持`,
						nil,
					)

					return rule.GetFirst(s)
				}),
		},
		&crawler.FieldItem{
			Name: "发起人历史发起的项目次数",
			Rule: ruler.NewCutStringRule(
				`<ul class="v-layout"><li>发起`,
				`</li>`,
				func(s string) string {
					rule := ruler.NewCutStringRule(
						`过`,
						`个项目`,
						nil,
					)

					return rule.GetFirst(s)
				}),
		},
		&crawler.FieldItem{
			Name: "发起人所在地",
			Rule: ruler.NewCutStringRule(
				`<span>所在地: `,
				`</span>`,
				nil,
			)},
		&crawler.FieldItem{
			Name: "项目结束时间",
			Rule: ruler.NewCutStringRule(
				`<p>若不能在 `,
				` 前筹集到`,
				nil,
			)},
		&crawler.FieldItem{
			Name: "浏览次数",
			Rule: ruler.NewCutStringRule(
				`<span class="h3plus views">`,
				`</span>`,
				func(s string) string {
					return strings.Replace(s, ",", "", -1)
				}),
		},

		&crawler.FieldItem{
			Name: "关注数",
			Rule: ruler.NewCutStringRule(
				`<span class="project-follower-count ml-5">`,
				`</span>`,
				func(s string) string {
					return strings.Replace(s, ",", "", -1)
				}),
		},

		&crawler.FieldItem{
			Name: "分享数",
			Rule: ruler.NewCutStringRule(
				`<span class="sns-share-num">`,
				`</span>`,
				func(s string) string {
					s = strings.TrimSpace(s)
					return strings.Replace(s, ",", "", -1)
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
			`http://www.qingju.com/projects/[0-9]+`,
			nil,
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
