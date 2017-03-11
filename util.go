package crawler

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
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
		return GBK2UTF8(string(respData)), nil
	}

	return string(respData), nil
}

func GetCookie(url string, key string) (val string, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("Error happens when create request to %s", url)
		return "", err
	}

	r, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Error happens when post to %s", url)
		return "", err
	}

	defer r.Body.Close()

	c, err := r.Cookie(key)
	if err != nil {
		log.Printf("Error happens when get cookie: %s", key)
		return "", err
	}

	return c.Value, nil
}

func GetWithCookie(url string, key, val string) (resp string, err error) {

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

func GBK2UTF8(gbkString string) string {
	dec := mahonia.NewDecoder("GBK")
	utfStr := dec.ConvertString(gbkString)
	return utfStr
}

func Unicode2UTF8(ustr string) string {
	var result string
	var coming bool
	for i := 0; i < len(ustr); i++ {
		switch ustr[i] {
		case '\\':
			coming = true
			continue
		case 'u':
			if coming {
				s := ustr[i+1 : i+5]
				i += 4
				t, err := strconv.ParseInt(s, 16, 32)
				if err != nil {
					result += s
					coming = false
					continue
				}
				coming = false
				result += fmt.Sprintf("%c", t)
			}
		default:
			result += ustr[i : i+1]
			coming = false
		}
	}

	return result
}
