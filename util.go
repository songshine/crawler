package crawler

import (
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/axgle/mahonia"
	"github.com/songshine/crawler/ruler"
)

// PostString make a POST HTTP request and encode a string data into ulr.
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

// Get makes a GET HTTP request from `url`.
func Get(url string) (resp string, err error) {
	r, err := http.Get(url)
	if err != nil {
		log.Printf("Error happens when get from %s, error %v", url, err)
		return "", err
	}

	defer r.Body.Close()
	respData, err := ioutil.ReadAll(r.Body)

	if err != nil {
		log.Printf("GET: Fail to parse response, error %v", err)
		return "", err
	}

	if strings.Contains(r.Header.Get("Content-Type"), "GBK") {
		// GBK => UTF-8
		return convertGBKToUTF8(string(respData)), nil
	}

	return string(respData), nil
}

// GetFromNextPage is a wrapper function to send a GET request and then
// perform`rule` on the response from GET.
func GetFromNextPage(url string, rule ruler.Interface) string {
	resp, err := Get(url)

	if err != nil {
		return ""
	}
	return rule.GetFirst(resp)
}

func convertGBKToUTF8(gbkString string) string {
	dec := mahonia.NewDecoder("gbk")
	return dec.ConvertString(gbkString)
}
