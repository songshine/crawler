package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"github.com/songshine/crawler"
)

const (
	StartCrawlerURL       = "https://z.jd.com/bigger/search.html"
	PaginationPostDataFmt = "status=&sort=&categoryId=%s&parentCategoryId=&sceneEnd=&productEnd=&keyword=&page=%d"
	ProjectRegexMatch     = `<a href="/project/details/[0-9]+.html"`
	ProjectURLFmt         = "https://z.jd.com/project/details/%s.html"
)

type project struct {
	category string
	id       string
	url      string
}

type projectData struct {
	key   string
	items map[string]string
}

var (
	Categories = map[string]string{
		"10": "科技",
	}

	TotalPagesOfCategory = map[string]int{
		"10": 2,
	}

	projectsChan = make(chan project, 500)
	resultsChan  = make(chan projectData, 200)
)

func fetchURLs() {
	var wg sync.WaitGroup
	wg.Add(len(Categories))
	for code, name := range Categories {
		totalPage := TotalPagesOfCategory[code]
		// one goroutine per category to fetch project urls
		go func(c, n string, total int) {
			for page := 1; page <= totalPage; page++ {
				fetchPageURLs(c, n, page)
			}
			wg.Done()
		}(code, name, totalPage)

	}
	wg.Wait()
	close(projectsChan)
}

func fetchPageURLs(code, category string, pageNum int) {
	postData := fmt.Sprintf(PaginationPostDataFmt, code, pageNum)
	body := ioutil.NopCloser(strings.NewReader(postData))
	req, err := http.NewRequest("POST", StartCrawlerURL, body)
	if err != nil {
		log.Printf("Error happens when create request to %s", StartCrawlerURL)
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Error happens when post to %s", StartCrawlerURL)
		return
	}

	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Fail to parse response")
		return
	}
	response := string(data)

	allProjects := getStringMatch(response, ProjectRegexMatch, true)
	for _, p := range allProjects {
		// get project id
		id := getStringMatch(p, `[0-9]+`, true)[0]
		projectsChan <- project{category: category, id: id, url: fmt.Sprintf(ProjectURLFmt, id)}
	}
}

func getStringMatch(content string, regex string, rmDup bool) []string {
	r := regexp.MustCompile(regex)
	matches := r.FindAllString(content, -1)
	if !rmDup {
		return matches
	}
	var result []string
	dupCheck := make(map[string]struct{})
	for _, m := range matches {
		if _, ok := dupCheck[m]; ok {
			continue
		}
		dupCheck[m] = struct{}{}
		result = append(result, m)
	}

	return result
}

func main() {
	// fetchURLs()
	go fetchURLs()
	i := 0
	for p := range projectsChan {
		fmt.Println(p, i)
		i++
	}

}
