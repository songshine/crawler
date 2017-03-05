package main

import (
	"fmt"

	"github.com/songshine/crawler/example/phantomjs/phantom"
)

func main() {
	p, err := phantom.Start()
	if err != nil {
		panic(err)
	}
	defer p.Exit() // Don't forget to kill phantomjs at some point.

	var (
		url      = "http://z.jd.com/project/details/74687.html"
		elemCond = `document.getElementById('focusCount').innerHTML != '0'`
		elem     = `document.getElementById('focusCount').innerHTML`
	)
	//jsCode := fmt.Sprintf(jsBody, url, elemCond, elem, timeoutMillis)

	//err = p.Run(fmt.Sprintf(jsBody, url, elemCond, elem, 2000), &result)

	fmt.Println(">>>>>>>>>>>>>>>>>>READY>>>>>>>>>>>>>>>>")
	//err = p.Run(fmt.Sprintf(jsBody, url, elemCond, elem, 1000), &result)
	for i := 0; i < 10; i++ {
		r, e := p.Run(fmt.Sprintf(jsBody, url, elemCond, elem, 2000))
		fmt.Println(">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>", r, e)
	}

	//fmt.Println(crawler.EvaluateJS(`http://z.jd.com/project/details/74687.html`, `document.getElementById('focusCount').innerHTML`, "(0)"))
}

var jsBody = `
 page.open('%s', function(status) {
        if (status !== 'success') {
            console.error("SH_RES" + " " + "NetworkError" + "\n");
        } else {
            waitFor(function() {
                    return page.evaluate(function() {
                        return %s;
                    })
            }, function(){
                    var result = page.evaluate(function() {
                        return %s;
                    });
					console.log("SH_RES" + " " + JSON.stringify(result) + "\n");	 
            }, %d) 
		} 

 });
 `
