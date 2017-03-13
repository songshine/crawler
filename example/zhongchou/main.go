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
	PageURLFmt = `http://www.zhongchou.com/browse/id-%s-si_c-p%d`
)

var (
	Categories = map[string]string{
		// "23": "公益",
		// "28":    "农业",
		// "16":    "出版",
		//"10001": "娱乐",
		"22": "艺术",
		"18": "其他",
	}

	StartPagesOfCategory = map[string]int{
		"23":    1,
		"28":    1,
		"16":    1,
		"10001": 1,
		"22":    1,
		"18":    1,
	}
	EndPageOfCategory = map[string]int{
		"23":    62,
		"28":    47,
		"16":    30,
		"10001": 35,
		"22":    33,
		"18":    55,
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
			Rule: ruler.NewXPathNodeRule(`//*[@id="move"]`, nil),
		},
		&crawler.FieldItem{
			Name: "众筹状态", ////*[@id="jlxqOuterBox"]/div/div[1]/div[2]/div[2]/div[2]/div[2]/span[1]
			Rule: ruler.NewXPathNodeRule(
				`//*[@id="jlxqOuterBox"]/div/div[1]/div[2]/div[2]/div[2]/div[2]/span[1]`,
				func(s string) string {
					if strings.Contains(s, "成功结束") {
						return "已成功"
					}

					return "众筹中"
				}),
		},
		&crawler.FieldItem{ //目标筹资<b>¥200,000</b></span> 目标筹资<b>¥10,000</b></span>
			Name: "目标金额",
			Rule: ruler.NewCutStringRule(`目标筹资<b>¥`, `</b></span>`, func(s string) string { return strings.Replace(s, ",", "", -1) }),
		},
		&crawler.FieldItem{
			Name: "实际筹资额",
			Rule: ruler.NewXPathNodeRule(`//*[@id="jlxqOuterBox"]/div/div[1]/div[2]/div[2]/div[1]/div[2]/p/span[1]`,
				func(s string) string {
					s = strings.Replace(s, ",", "", -1)
					return strings.TrimPrefix(s, "¥")
				}),
		},
		&crawler.FieldItem{
			Name: "项目支持人数",
			Rule: ruler.NewXPathNodeRule(`//*[@id="jlxqOuterBox"]/div/div[1]/div[2]/div[2]/div[1]/div[1]/p/span[1]`, func(s string) string { return strings.Replace(s, ",", "", -1) }),
		}, //
		&crawler.FieldItem{
			Name: "项目回报总类",
			Rule: ruler.NewCutStringRule(`<div class="zcje_ItemBox">`, `<!-- 支持金额列表 end -->`, func(s string) string {
				return strconv.Itoa(strings.Count(s, `class="zcje_ItemBox"`))
			}),
		},
		&crawler.FieldItem{
			Name: "最低投资额",
			Rule: ruler.NewCutStringRule(`<div class="zcje_ItemBox">`, `<!-- 支持金额列表 end -->`, func(s string) string {
				rule := ruler.NewCutStringRule(
					`<b>¥`,
					`</b>`,
					func(s string) string {
						return strings.TrimSpace(s)
					},
				)
				prices := rule.Get(s, true)
				minPrice, found := 0.0, false
				for _, p := range prices {
					pi, err := strconv.ParseFloat(p, 64)
					if err == nil && (!found || pi < minPrice) {
						found = true
						minPrice = pi
					}
				}
				return strconv.FormatFloat(minPrice, 'f', 2, 64)

			}),
		},

		&crawler.FieldItem{
			Name: "项目图片数量",
			Rule: ruler.NewCutStringRule(`<!-- 详情页 -->`, `<!-- 分享 s-->`, func(s string) string { return strconv.Itoa(strings.Count(s, `<img `)) }),
		},
		&crawler.FieldItem{
			Name: "项目更新数",
			Rule: ruler.NewCutStringRule(`项目更新（<b>`, `</b>`, nil),
		},
		&crawler.FieldItem{
			Name: "评论数",
			Rule: ruler.NewCutStringRule(`评论（<b>`, `</b>）`, nil),
		},
		&crawler.FieldItem{
			Name: "是否制作视频",
			Rule: ruler.NewNooptRule(func(s string) string {
				if strings.Contains(s, `<div class="play-box">`) {
					return "是"
				}
				return "否"
			}),
		},
		&crawler.FieldItem{
			Name: "无私支持数",
			Rule: ruler.NewCutStringRule(`无私支持<b>`, ` 人</b>`, nil),
		},
		&crawler.FieldItem{
			Name: "最低投资支持人数",
			Rule: ruler.NewCutStringRule(`<div class="zcje_ItemBox">`, `<!-- 支持金额列表 end -->`,
				func(s string) string {
					rule := ruler.NewCutStringRule(
						`<h3 class="zcje_h3"><b>¥`,
						`</h3>`,
						func(s string) string {
							ss := strings.NewReplacer("</b>", "&", "／限", "&", ",", "", "人", "", " ", "")
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
							if ps[1] == "<em>已满额</em>" && len(ps) > 2 {
								minSupport = ps[2]
								continue
							}
							minSupport = ps[1]

						}

					}

					return minSupport
				}),
		},
		&crawler.FieldItem{
			Name: "预计回报发送时间", //*[@id="right"]/div/div/div[1]/div/div[2]/div[3]/p/b
			Rule: ruler.NewCutStringRule(
				`<p>预计回报发送时间：<b>`,
				`<!--项目结束后立即回报项目成功结束后天内-->`,
				func(s string) string {
					if s != "" {
						return strings.TrimSpace(s)
					}
					return "不确定"
				}),
		},

		&crawler.FieldItem{
			Name: "发起人所在地",
			Rule: ruler.NewNooptRule(
				func(s string) string {
					ruleP := ruler.NewXPathNodeRule(
						`//*[@id="jlxqOuterBox"]/div/div[1]/div[2]/div[2]/div[3]/div[2]/div/span[2]/a`,
						func(s string) string {
							return strings.TrimSpace(s)
						},
					)
					ruleC := ruler.NewXPathNodeRule(
						`//*[@id="jlxqOuterBox"]/div/div[1]/div[2]/div[2]/div[3]/div[2]/div/span[2]/a[2]`,
						func(s string) string {
							return strings.TrimSpace(s)
						},
					)

					pri := ruleP.GetFirst(s)
					cit := ruleC.GetFirst(s)
					de := " "
					if pri == "" || cit == "" {
						de = ""
					}

					return pri + de + cit
				}),
		},
		&crawler.FieldItem{
			Name: "项目结束时间",
			Rule: ruler.NewXPathNodeRule(
				`//*[@id="jlxqOuterBox"]/div/div[1]/div[2]/div[2]/div[2]/div[2]/span[1]/b`,
				func(s string) string {
					s = strings.TrimSpace(s)
					if s != "" {
						return s
					}

					return "已经结束"
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
			`http://www.zhongchou.com/deal-show/id-[0-9]+`,
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
