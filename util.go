package crawler

import (
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/songshine/crawler/ruler"
)

func PostString(url, data string) (resp string, err error) {
	body := ioutil.NopCloser(strings.NewReader(data))
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		log.Printf("Error happens when create request to %s", url)
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	r, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Error happens when post to %s", url)
		return "", err
	}

	defer r.Body.Close()

	respData, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("POST: Fail to parse response")
		return "", err
	}
	return string(respData), nil
}

func Get(url string) (resp string, err error) {
	r, err := http.Get(url)
	if err != nil {
		log.Printf("Error happens when get from %s", url)
		return "", err
	}

	defer r.Body.Close()
	respData, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("GET: Fail to parse response")
		return "", err
	}

	return string(respData), nil
}

func GetFromNextPage(url string, rule ruler.Interface) string {
	resp, err := Get(url)

	if err != nil {
		return ""
	}
	return rule.GetFirst(resp)
}
