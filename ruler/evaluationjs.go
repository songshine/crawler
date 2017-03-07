package ruler

import (
	"fmt"

	"log"

	"github.com/songshine/crawler/phantom"
)

func NewEvaluationJSRule(elemJS, checkJS string, timeoutInMillis int, transFunc transStringFunc) Interface {
	r := &evaluationJSRule{
		elemJS:          elemJS,
		checkJS:         checkJS,
		timeoutInMillis: timeoutInMillis,
		trans:           transFunc,
	}
	return r
}

// implement Interface
type evaluationJSRule struct {
	elemJS, checkJS string
	timeoutInMillis int
	trans           transStringFunc
}

func (r *evaluationJSRule) Get(url string, distinct bool) []string {
	// ONLY SUPPORT TO RETURN ONE VALUE NOW.
	return r.trans.transStringSlice([]string{evaluateJS(url, r.elemJS, r.checkJS, r.timeoutInMillis)})
}

func (r *evaluationJSRule) GetFirst(url string) string {
	return r.trans.transString(evaluateJS(url, r.elemJS, r.checkJS, r.timeoutInMillis))
}

func evaluateJS(url, elemJs string, checkJS string, timeoutInMillis int) string {
	log.Println(">>>>>>>>>>>>>>>>>>>>>>>> evaluateJS")
	js := fmt.Sprintf(jsBody, url, checkJS, elemJs, timeoutInMillis)
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
