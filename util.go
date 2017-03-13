package crawler

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/axgle/mahonia"
	"github.com/songshine/crawler/ruler"
)

func getTransport(addr string) (transport *http.Transport) {
	url := url.URL{}
	urlProxy, _ := url.Parse(addr)
	transport = &http.Transport{Proxy: http.ProxyURL(urlProxy)}
	return
}

// PostString make a POST HTTP request and encode a string data into ulr.
func PostString(url, data string) (resp string, err error) {
	body := ioutil.NopCloser(strings.NewReader(data))
	req, err := http.NewRequest("POST", url, body)
	//req.Header.Set("User-Agent", GetRandomUserAgent())
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
	req, err := http.NewRequest("GET", url, nil)
	//req.Header.Set("User-Agent", GetRandomUserAgent())
	if err != nil {
		log.Printf("Error happens when create request to %s", url)
		return "", err
	}
	/*transport := getTransport("https://agentwho.rocks/S6w8M73.pac")
	client := &http.Client{Transport: transport}*/
	r, err := http.DefaultClient.Do(req)
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

func GetWithClient(url string, client *http.Client) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	//req.Header.Set("User-Agent", GetRandomUserAgent())
	if err != nil {
		log.Printf("Error happens when create request to %s", url)
		return "", err
	}

	r, err := client.Do(req)
	if err != nil {
		log.Printf("Error happens when get from %s, error: %v", url, err)
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

// GetFromNextPage is a wrapper function to send a GET request and then
// perform`rule` on the response from GET.
func GetFromNextPage(url string, rule ruler.Interface) string {
	resp, err := Get(url)

	if err != nil {
		return ""
	}
	return rule.GetFirst(resp)
}

func GetFromNextPageWithClient(url string, rule ruler.Interface, client *http.Client) string {
	resp, err := GetWithClient(url, client)
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

var userAgent = [...]string{"Mozilla/5.0 (compatible, MSIE 10.0, Windows NT, DigExt)",
	"Mozilla/4.0 (compatible, MSIE 7.0, Windows NT 5.1, 360SE)",
	"Mozilla/4.0 (compatible, MSIE 8.0, Windows NT 6.0, Trident/4.0)",
	"Mozilla/5.0 (compatible, MSIE 9.0, Windows NT 6.1, Trident/5.0,",
	"Opera/9.80 (Windows NT 6.1, U, en) Presto/2.8.131 Version/11.11",
	"Mozilla/4.0 (compatible, MSIE 7.0, Windows NT 5.1, TencentTraveler 4.0)",
	"Mozilla/5.0 (Windows, U, Windows NT 6.1, en-us) AppleWebKit/534.50 (KHTML, like Gecko) Version/5.1 Safari/534.50",
	"Mozilla/5.0 (Macintosh, Intel Mac OS X 10_7_0) AppleWebKit/535.11 (KHTML, like Gecko) Chrome/17.0.963.56 Safari/535.11",
	"Mozilla/5.0 (Macintosh, U, Intel Mac OS X 10_6_8, en-us) AppleWebKit/534.50 (KHTML, like Gecko) Version/5.1 Safari/534.50",
	"Mozilla/5.0 (Linux, U, Android 3.0, en-us, Xoom Build/HRI39) AppleWebKit/534.13 (KHTML, like Gecko) Version/4.0 Safari/534.13",
	"Mozilla/5.0 (iPad, U, CPU OS 4_3_3 like Mac OS X, en-us) AppleWebKit/533.17.9 (KHTML, like Gecko) Version/5.0.2 Mobile/8J2 Safari/6533.18.5",
	"Mozilla/4.0 (compatible, MSIE 7.0, Windows NT 5.1, Trident/4.0, SE 2.X MetaSr 1.0, SE 2.X MetaSr 1.0, .NET CLR 2.0.50727, SE 2.X MetaSr 1.0)",
	"Mozilla/5.0 (iPhone, U, CPU iPhone OS 4_3_3 like Mac OS X, en-us) AppleWebKit/533.17.9 (KHTML, like Gecko) Version/5.0.2 Mobile/8J2 Safari/6533.18.5",
	"MQQBrowser/26 Mozilla/5.0 (Linux, U, Android 2.3.7, zh-cn, MB200 Build/GRJ22, CyanogenMod-7) AppleWebKit/533.1 (KHTML, like Gecko) Version/4.0 Mobile Safari/533.1"}

var r = rand.New(rand.NewSource(time.Now().UnixNano()))

func GetRandomUserAgent() string {
	return userAgent[r.Intn(len(userAgent))]
}
