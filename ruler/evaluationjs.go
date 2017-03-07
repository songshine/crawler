package ruler

import (
	"fmt"

	"github.com/songshine/crawler/phantom"
	"github.com/songshine/crawler/pool"
)

const maxRuningPhantomjs = 2

var globalPhantomjsPool pool.Interface

func init() {
	globalPhantomjsPool = pool.New(maxRuningPhantomjs)
}

func NewEvaluationJSRule(script, exclude string, timeoutInMillis int, transFunc transStringFunc) Interface {
	r := &evaluationJSRule{
		script:          script,
		exclude:         exclude,
		timeoutInMillis: timeoutInMillis,
		trans:           transFunc,
	}
	return r
}

// implement ExtractStringRuler
type evaluationJSRule struct {
	script, exclude string
	timeoutInMillis int
	trans           transStringFunc
}

func (r *evaluationJSRule) Get(url string, distinct bool) []string {
	// ONLY SUPPORT TO RETURN ONE VALUE NOW.
	return r.trans.transStringSlice([]string{evaluateJS(url, r.script, r.exclude, r.timeoutInMillis)})
}

func (r *evaluationJSRule) GetFirst(url string) string {
	return r.trans.transString(evaluateJS(url, r.script, r.exclude, r.timeoutInMillis))
}

func evaluateJS(url, elemJs string, excludeJs string, timeoutInMillis int) string {
	js := fmt.Sprintf(jsBody, url, excludeJs, elemJs, timeoutInMillis)
	p := phantom.Take()
	defer phantom.Return(p)
	result, err := p.Run(js)
	if err != nil {
		return err.Error()
	}

	return result
}

var jsBody = `
var page = webpage.create();
page.onResourceRequested = function(requestData, request) {
    if ((/http:\/\/.+?\.css/gi).test(requestData['url']) || requestData['Content-Type'] == 'text/css') {
        request.abort();
    }
};
page.open('%s', function(status) {
        if (status !== 'success') {
			page.close();
            console.error("SH_RES" + " " + "NetworkError" + "\n");
        } else {
			var check = function() {
				return page.evaluate(function() {
						return %s;
				});
            };
			var done = function() {
				var result = page.evaluate(function() {
                    	return %s;
                });								
				console.log("SH_RES" + " " + result + "\n");	
				page.close();			
			};
			if (check()) {
				done();
				setTimeout(captureInput, 0);			
			} else {
				waitFor(check, done, %d);
			}   
		} 
 });
 `
