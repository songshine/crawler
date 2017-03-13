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
	PageURLFmt = `http://www.demohour.com/projects?attribute=most_funded&category=%s&page=%d&total_results=10000`
)

var (
	Categories = map[string]string{
		"927258": "电视电影",
		"927259": "广告会展",
		"927219": "通讯数码",
		"927151": "家居生活",
		"927158": "智能穿戴",
		"927218": "影音娱乐",
		"927162": "出行定位",
		"927256": "文化艺术",
		"927257": "饮食文化",
	}

	StartPagesOfCategory = map[string]int{
		"927258": 1,
		"927259": 1,
		"927219": 1,
		"927151": 1,
		"927158": 1,
		"927218": 1,
		"927162": 1,
		"927256": 1,
		"927257": 1,
	}
	EndPageOfCategory = map[string]int{
		"927258": 1,
		"927259": 1,
		"927219": 4,
		"927151": 10,
		"927158": 4,
		"927218": 4,
		"927162": 3,
		"927256": 2,
		"927257": 2,
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
				`<div class="c40">`,
				`</div>`,
				nil),
		},
		&crawler.FieldItem{
			Name: "众筹状态",
			Rule: ruler.NewCutStringRule(
				`<div class="c4">
<span class="c6">`,
				`</span>`,
				nil,
			),
		},
		&crawler.FieldItem{
			Name: "目标金额",
			Rule: ruler.NewCutStringRule(`目标<b>¥</b>`, `<span>`, func(s string) string { return strings.Replace(s, ",", "", -1) }),
		},
		&crawler.FieldItem{
			Name: "实际筹资额",
			Rule: ruler.NewCutStringRule(`<strong><b>¥</b>`, `</strong>`,
				func(s string) string {
					return strings.Replace(s, ",", "", -1)
				}),
		},
		&crawler.FieldItem{
			Name: "项目支持人数",
			Rule: ruler.NewCutStringRule(`/strong>
<span>`,
				`人支持</span>`,
				func(s string) string { return strings.Replace(s, ",", "", -1) }),
		},
		&crawler.FieldItem{
			Name: "项目回报总类",
			Rule: ruler.NewNooptRule(func(s string) string {
				return strconv.Itoa(strings.Count(s, `<dl class="c130">
<dt><b>¥`))
			}),
		},
		&crawler.FieldItem{
			Name: "最低投资额",
			Rule: ruler.NewNooptRule(func(s string) string {
				rule := ruler.NewCutStringRule(
					`<dl class="c130">
<dt><b>¥</b>`,
					`
<span>`,
					func(s string) string {
						s = strings.TrimSpace(s)
						return strings.Replace(s, ",", "", -1)
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
			Rule: ruler.NewNooptRule(func(s string) string { return strconv.Itoa(strings.Count(s, `<img src`)) }),
		},
		&crawler.FieldItem{
			Name: "公告数",
			Rule: ruler.NewXPathNodeRule(`//*[@id="tab_posts"]/a/span`,
				func(s string) string {
					if s == "" {
						s = "0"
					}
					return strings.Replace(s, ",", "", -1)
				}),
		},
		&crawler.FieldItem{
			Name: "评论数",
			Rule: ruler.NewXPathNodeRule(`//*[@id="tab_reviews"]/a/span`,
				func(s string) string {
					if s == "" {
						s = "0"
					}
					return strings.Replace(s, ",", "", -1)
				}),
		},
		&crawler.FieldItem{
			Name: "是否制作视频",
			Rule: ruler.NewNooptRule(func(s string) string {
				if strings.Contains(s, `<embed wmode="opaque"`) {
					return "是"
				}
				return "否"
			}),
		},

		&crawler.FieldItem{
			Name: "预计回报发送时间",
			Rule: ruler.NewCutStringRule(
				`<p class="c1241">回报时间：`,
				`</p>`,
				func(s string) string {
					if s != "" {
						return strings.TrimSpace(s)
					}
					return "不确定"
				}),
		},
		&crawler.FieldItem{
			Name: "粉丝数",
			Rule: ruler.NewCutStringRule(
				`<dl class="project-initiator">
<dd class="c8"><a href="`,
				`" target`,
				func(s string) string {
					if s == "" {
						return ""
					}

					url := "http://www.demohour.com" + s
					rule := ruler.NewXPathNodeRule(
						`/html/body/div[2]/div[1]/div[2]/div[1]/a[2]/strong`,
						nil,
					)

					ret := crawler.GetFromNextPage(url, rule)
					if ret == "" {
						return "0"
					}
					return ret
				}),
		},
		&crawler.FieldItem{
			Name: "发起人支持的项目数",
			Rule: ruler.NewCutStringRule(
				`<dl class="project-initiator">
<dd class="c8"><a href="`,
				`" target`,
				func(s string) string {
					if s == "" {
						return ""
					}

					url := "http://www.demohour.com" + s
					rule := ruler.NewXPathNodeRule(
						`/html/body/div[2]/div[1]/div[2]/div[1]/a[4]/strong`,
						nil,
					)

					return crawler.GetFromNextPage(url, rule)
				}),
		},
		&crawler.FieldItem{
			Name: "发起人历史发起的项目次数",
			Rule: ruler.NewCutStringRule(
				`<dl class="project-initiator">
<dd class="c8"><a href="`,
				`" target`,
				func(s string) string {
					if s == "" {
						return ""
					}

					url := "http://www.demohour.com" + s
					rule := ruler.NewXPathNodeRule(
						`/html/body/div[2]/div[1]/div[2]/div[1]/a[3]/strong`,
						nil,
					)

					return crawler.GetFromNextPage(url, rule)
				}),
		},
		&crawler.FieldItem{
			Name: "支持次数",
			Rule: ruler.NewXPathNodeRule(`//*[@id="tab_backers"]/a/span`, func(s string) string {
				if s == "" {
					return "0"
				}
				return strings.Replace(s, ",", "", -1)
			}),
		},
		&crawler.FieldItem{
			Name: "最低投资支持人数",
			Rule: ruler.NewNooptRule(
				func(s string) string {
					rule := ruler.NewCutStringRule(
						`<dl class="c130">
<dt><b>¥</b>`,
						`位参与`,
						func(s string) string {
							ss := strings.NewReplacer("<span>（", "&", ",", "", " ", "", "\n", "")
							return ss.Replace(strings.TrimSpace(s))
						},
					)
					pairs := rule.Get(s, false)
					var (
						minSupport string
						minPrice   float64
						found      bool
					)

					for _, p := range pairs {
						ps := strings.Split(p, "&")
						if len(ps) < 2 {
							continue
						}

						pi, err := strconv.ParseFloat(ps[0], 64)
						if err == nil && (!found || pi < minPrice) {
							found = true
							minPrice = pi
							minSupport = ps[1]

						}

					}

					return minSupport
				}),
		},

		&crawler.FieldItem{
			Name: "发起人所在地",
			Rule: ruler.NewCutStringRule(
				`<dl class="project-initiator">
<dd class="c8"><a href="`,
				`" target`,
				func(s string) string {
					if s == "" {
						return ""
					}

					url := "http://www.demohour.com" + s
					rule := ruler.NewXPathNodeRule(
						`/html/body/div[2]/div[1]/div[1]/dl/dd[4]/p`,
						nil,
					)

					return crawler.GetFromNextPage(url, rule)
				}),
		},
		&crawler.FieldItem{
			Name: "项目结束时间",
			Rule: ruler.NewNooptRule(

				func(s string) string {
					avaiXPathes := []string{
						`/html/body/div[4]/div[2]/div[2]/div[1]/div/div[2]/span`,
						`/html/body/div[4]/div[1]/div[2]/div[1]/div[1]/div[2]/span`,
					}
					ret := "已经结束"
					for _, p := range avaiXPathes {
						rule := ruler.NewXPathNodeRule(
							p,
							nil,
						)

						t := rule.GetFirst(s)
						if t != "" {
							ret = t
							break
						}
					}
					return ret
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
			`/projects/[0-9]+`,
			func(s string) string {
				return `http://www.demohour.com` + s
			})
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
