package ruler

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"text/template"
	"time"

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
	return r.trans.transStringSlice([]string{evaluateJS2(url, r.script, r.exclude, r.timeoutInMillis)})
}

func (r *evaluationJSRule) GetFirst(url string) string {
	return r.trans.transString(evaluateJS2(url, r.script, r.exclude, r.timeoutInMillis))
}

func evaluateJS2(url, elemJs string, excludeJs string, timeoutInMillis int) string {
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
 page.open('%s', function(status) {
        if (status !== 'success') {
			page.release();
            console.error("SH_RES" + " " + "NetworkError" + "\n");
			setTimeout(captureInput, 0);
        } else {
            waitFor(function() {
                    return page.evaluate(function() {
                        return %s;
                    })
            }, function(){
                    var result = page.evaluate(function() {
                        return %s;
                    });					
					console.log("SH_RES" + " " + result + "\n");
					page.release()
            }, %d) 
		} 
 });
 `

func evaluateJS(url, script string, exclude string, timeoutInMillis int) string {
	data := struct {
		URL     string
		Script  string
		Exclude string
		Timeout int
	}{url, script, exclude, timeoutInMillis}

	tmpl, err := template.New("evaluatejs").Parse(evaluateJSTmpl)
	if err != nil {
		return ""
	}

	tmpFileName := fmt.Sprintf("%d.js", time.Now().Nanosecond())
	tmpFile, _ := os.Create(tmpFileName)
	defer os.Remove(tmpFileName)

	tmpl.Execute(tmpFile, data)
	tmpFile.Close()

	ticket := globalPhantomjsPool.Take()
	defer globalPhantomjsPool.Return(ticket)

	cmd := exec.Command("bin/phantomjs", tmpFileName)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return ""
	}

	done := make(chan string, 1)
	go func() {
		err = cmd.Start()
		result, err := ioutil.ReadAll(stdout)
		if err != nil {
			done <- ""
			return
		}

		done <- string(result)
	}()

	select {
	case <-time.After(time.Millisecond * time.Duration((timeoutInMillis + 3000))):
		cmd.Process.Kill()
		return ""
	case r := <-done:
		return r
	}
}

var evaluateJSTmpl = `
"use strict";
function waitFor(testFx, onReady, timeOutMillis) {
    var maxtimeOutMillis = timeOutMillis ? timeOutMillis : {{.Timeout}},
        start = new Date().getTime(),
        condition = false,
        interval = setInterval(function() {
            if ( (new Date().getTime() - start < maxtimeOutMillis) && !condition ) {
                condition = (typeof(testFx) === "string" ? eval(testFx) : testFx()); 
            } else {
                if(!condition) {
                    console.log("timeout")
                    phantom.exit(1);
                } else {                
                    typeof(onReady) === "string" ? eval(onReady) : onReady();
                    clearInterval(interval);
                }
            }
        }, 500);
};

var page = require('webpage').create();
page.open('{{.URL}}', function(status) {
    if (status == 'success') {
    	waitFor(function() {
            return page.evaluate(function() {
                return {{.Script}} !== '{{.Exclude}}';
            })
		}, function(){
				var val = page.evaluate(function() {
					return {{.Script}};
				});
				console.log(val);
				phantom.exit();
		}) 
  	}
});
`
