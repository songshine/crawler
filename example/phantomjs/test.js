"use strict";
var system = require('system');
var page = require('webpage').create();
(function() {
    function captureInput() {
        var lines = [];
        var l = system.stdin.readLine();
        while (l !== 'END' && l !== 'END\n') {
            lines.push(l);
            l = system.stdin.readLine();
        }
        var command = lines.splice(0, 1)[0];
        if (command === 'EVAL' || command === 'EVAL\n') {
        try {
            eval.call(this, lines.join('\n'));
        } catch (ex) {
            system.stderr.writeLine("Error during EVAL of" + lines.join('\n'));
            setTimeout(captureInput, 0);
        }
        } else if (command === 'RUN' || command === 'RUN\n') {
            try {
                eval(lines.join('\n'));
            }catch (ex) {
                system.stderr.writeLine("SH_RES" + " " + JSON.stringify(ex) + "\n");
                setTimeout(captureInput, 0);                
            }  
                
        } else {
            system.stderr.writeLine("Invalid command:<" + command+">");
            setTimeout(captureInput, 0);
        }        
    }
    function waitFor(testFx, onReady, timeOutMillis) {
        var maxtimeOutMillis = timeOutMillis ? timeOutMillis : 3000,
        start = new Date().getTime(),
        condition = false,
        interval = setInterval(function() {
            if ( (new Date().getTime() - start < maxtimeOutMillis) && !condition ) {
                condition = (typeof(testFx) === "string" ? eval(testFx) : testFx()); 
            } else {
                if(!condition) {
                    system.stderr.writeLine("SH_RES" + " " + "TimoutError" + "\n");
                    
                } else {                        
                    typeof(onReady) === "string" ? eval(onReady) : onReady();
                    clearInterval(interval);
                }
                setTimeout(captureInput, 0);
            }
        }, 100);
    };
    setTimeout(captureInput, 0);
}())