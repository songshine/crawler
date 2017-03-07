package phantom

import (
	"bufio"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
)

const (
	ResponsePrefix = "SH_RES"

	MaxPhantomInstance = 3

	MaxTimeoutSecond = 5
)

var (
	pool            *phantomPool
	wrapperFileName string
)

type phantomPool struct {
	max int
	ps  chan *Phantom
}

func init() {
	var err error
	wrapperFileName, err = createWrapperFile()
	log.Printf(">>> Wrapper file name: %s", wrapperFileName)
	if err != nil {
		log.Printf("Create wrapper file failed")
		return
	}

	pool = &phantomPool{
		max: MaxPhantomInstance,
		ps:  make(chan *Phantom, MaxPhantomInstance),
	}

	for i := 0; i < pool.max; i++ {
		p, err := start("--load-images=no", "--disk-cache=yes")
		if err != nil {
			log.Printf("Start phantomjs failed, error: %s \n", err)
			continue
		}
		pool.ps <- p
	}
}

// Phantom represents a process of Phantomjs.
type Phantom struct {
	cmd     *exec.Cmd
	in      io.WriteCloser
	out     io.ReadCloser
	errout  io.ReadCloser
	resChan chan string
	errChan chan error
}

// Take takes a Phantom instance from Phantom pool.
func Take() *Phantom {
	return <-pool.ps
}

// Return gives back a Phantom instance to Phantom pool.
func Return(p *Phantom) {
	pool.ps <- p
}

// Exit exits all Phantom instances in pool.
func Exit() {
	close(pool.ps)
	for p := range pool.ps {
		err := p.exit()
		if err != nil {
			log.Printf("PhantomJS exit failed, error %v \n", err)
		}
	}
}

func start(scriptPath string, args ...string) (*Phantom, error) {
	args = append(args, wrapperFileName)
	cmd := exec.Command("phantomjs", args...)

	inPipe, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	outPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	errPipe, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	p := Phantom{
		cmd:     cmd,
		in:      inPipe,
		out:     outPipe,
		errout:  errPipe,
		resChan: make(chan string), // need buffer?
		errChan: make(chan error),  // need buffer?
	}
	err = cmd.Start()

	if err != nil {
		return nil, err
	}

	p.startReadStd()
	return &p, nil
}

func (p *Phantom) startReadStd() {
	go func() {
		scannerOut := bufio.NewScanner(p.out)
		for scannerOut.Scan() {
			line := scannerOut.Text()
			parts := strings.SplitN(line, " ", 2)
			if strings.HasPrefix(line, ResponsePrefix) {
				p.resChan <- parts[1]
				continue
			}
			log.Printf("INFO LOG %s\n", line)

		}
	}()
	go func() {
		scannerErrorOut := bufio.NewScanner(p.errout)
		for scannerErrorOut.Scan() {
			line := scannerErrorOut.Text()
			parts := strings.SplitN(line, " ", 2)
			if strings.HasPrefix(line, ResponsePrefix) {
				p.errChan <- errors.New(parts[1])
				continue
			}
			log.Printf("ERROR LOG %s\n", line)
		}
	}()
}

func (p *Phantom) exit() error {
	err := p.load("phantom.exit();")
	if err != nil {
		return err
	}

	p.in.Close()
	p.out.Close()
	p.errout.Close()

	err = p.cmd.Wait()
	if err != nil {
		return err
	}

	return nil
}

// Run sends JavaScript script into Phantom stdin, then wait result from stdout and stderr
func (p *Phantom) Run(jsScript string) (string, error) {
	err := p.sendLine("RUN", jsScript, "END")
	if err != nil {
		return "", err
	}

	select {
	case text := <-p.resChan:
		return text, nil
	case err := <-p.errChan:
		return "", err
	}
}

func (p *Phantom) load(jsCode string) error {
	return p.sendLine("EVAL", jsCode, "END")
}

func (p *Phantom) sendLine(lines ...string) error {
	for _, l := range lines {
		_, err := io.WriteString(p.in, l+"\n")
		if err != nil {
			return errors.New("Cannot Send: `" + l + "`" + err.Error())
		}
	}
	return nil
}

func createWrapperFile() (string, error) {
	wrapper, err := ioutil.TempFile("", "go-phantom-wrapper")
	if err != nil {
		return "", err
	}
	defer wrapper.Close()

	err = ioutil.WriteFile(wrapper.Name(), []byte(jsEntry), os.ModeType)
	if err != nil {
		return "", err
	}

	return wrapper.Name(), nil
}

var jsEntry = `
var system = require('system');
var webpage = require('webpage');

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
                system.stderr.writeLine("Error" + " " + JSON.stringify(ex) + "\n");
                system.stderr.writeLine("SH_RES" + " " + "UnknownError" + "\n");
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
        condition = false;
        var interval = setInterval(function() {
            if ( (new Date().getTime() - start < maxtimeOutMillis) && !condition ) {
                condition = (typeof(testFx) === "string" ? eval(testFx) : testFx()); 
            } else {
                if(!condition) {
                    system.stderr.writeLine("SH_RES" + " " + "TimoutError" + "\n");
                } else {      
                    typeof(onReady) === "string" ? eval(onReady) : onReady(); 	                  
                }
                clearInterval(interval);  
                setTimeout(captureInput, 0);
            }
        }, 250);
    };
    
    setTimeout(captureInput, 0);   
}());
`
