package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
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
	var tbToken string
	cookieJar, _ := cookiejar.New(nil)
	httpClient := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return nil
		},
		Jar: cookieJar,
	}
	testURL := `https://izhongchou.taobao.com/dreamdetail.htm?id=20058471`
	crawler.GetWithClient(testURL, httpClient)

	testReq, _ := http.NewRequest("GET", testURL, nil)
	cookies := httpClient.Jar.Cookies(testReq.URL)
	for _, c := range cookies {
		if c.Name == "_tb_token_" {
			tbToken = c.Value
			break
		}
	}
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
				url := fmt.Sprintf(`https://izhongchou.taobao.com/dream/ajax/get_project_feeds.htm?project_id=%s`, s)
				if tbToken != "" {
					url += fmt.Sprintf("&_tb_token_=%s", tbToken)
				}
				res, err := crawler.GetWithClient(url, httpClient)
				if err != nil {
					return ""
				}
				return strconv.Itoa(strings.Count(res, `"feed_id":`))
			}),
		},
		&crawler.FieldItem{
			Name: "喜欢数",
			Rule: ruler.NewCutStringRule(`"focus_count":"`, `",`, nil),
		},
		&crawler.FieldItem{
			Name: "项目结束时间",
			Rule: ruler.NewCutStringRule(`"end_date": "`, `",`, nil),
		},
		&crawler.FieldItem{
			Name: "是否制作视频",
			Rule: ruler.NewCutStringRule(`"video": "`, `",`,
				func(s string) string {
					if s == "" {
						return "否"
					}
					return "是"
				}),
		},
		&crawler.FieldItem{
			Name: "最低投资支持人数",
			Rule: ruler.NewCutStringRule(`"items":`, `]`,
				func(s string) string {
					rulePrice := ruler.NewCutStringRule(
						`"price": "`,
						`"`,
						func(s string) string {
							return strings.TrimSpace(s)
						},
					)
					prices := rulePrice.Get(s, false)

					rulePerson := ruler.NewCutStringRule(
						`"support_person": "`,
						`"`,
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
		&crawler.FieldItem{
			Name: "预计回报发送时间",
			Rule: ruler.NewCutStringRule(`"make_days": "`, `",`, nil),
		},
		&crawler.FieldItem{
			Name: "发起人是淘宝还是天猫店铺(描述,服务,物流评价)",
			Rule: ruler.NewCutStringRule(`"shopId": "`, `",`, func(s string) string {
				shopURL := fmt.Sprintf("https://shop%s.taobao.com", s)
				var desc string
				ruleNoop := ruler.NewNooptRule(func(s string) string {
					if strings.Contains(s, `title="天猫Tmall.com"`) {
						return "天猫"
					}

					return "淘宝"
				})
				desc = crawler.GetFromNextPageWithClient(shopURL, ruleNoop, httpClient)

				rateFunc := func(ss []string) string {

					for _, x := range ss {
						rule := ruler.NewXPathNodeRule(
							x,
							func(s string) string {
								return strings.TrimSpace(s)
							},
						)

						rate := crawler.GetFromNextPageWithClient(shopURL, rule, httpClient)
						if rate != "" {
							return rate
						}
					}

					return ""
				}
				availXPath1 := []string{
					`//*[@id="J_ShopRate2"]/ul/li[1]/em`,
					`//*[@id="header-content"]/div[2]/div[4]/div[2]/div[2]/ul/li[1]/em`,
					`//*[@id="shop-info"]/div[2]/div[1]/div[2]/span`,
				}

				availXPath2 := []string{
					`//*[@id="J_ShopRate2"]/ul/li[1]/em`,
					`//*[@id="header-content"]/div[2]/div[4]/div[2]/div[2]/ul/li[1]/em`,
					`//*[@id="shop-info"]/div[2]/div[1]/div[2]/span`,
				}

				availXPath3 := []string{
					`//*[@id="J_ShopRate2"]/ul/li[2]/em`,
					`//*[@id="header-content"]/div[2]/div[4]/div[2]/div[2]/ul/li[2]/em`,
					`//*[@id="shop-info"]/div[2]/div[2]/div[2]/span`,
				}
				var allRates string
				allRates += rateFunc(availXPath1) + ","
				allRates += rateFunc(availXPath2) + ","
				allRates += rateFunc(availXPath3)

				return desc + "-" + allRates
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
