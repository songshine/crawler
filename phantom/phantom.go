package phantom

import (
	"bufio"
	"errors"
	"io"
	"log"
	"os/exec"
	"strings"
)

const MaxPhantomInstance = 3
const MaxTimeoutSecond = 5

var pool *phantomPool

type phantomPool struct {
	max int
	ps  chan *Phantom
}

func init() {
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

func Take() *Phantom {
	return <-pool.ps
}

func Return(p *Phantom) {
	pool.ps <- p
}

func Exit() {
	close(pool.ps)
	for p := range pool.ps {
		err := p.exit()
		if err != nil {
			log.Printf("PhantomJS exit failed, error %v \n", err)
		}
	}
}

type Phantom struct {
	cmd     *exec.Cmd
	in      io.WriteCloser
	out     io.ReadCloser
	errout  io.ReadCloser
	resChan chan string
	errChan chan error
}

func start(args ...string) (*Phantom, error) {
	args = append(args, "data/wrapper.js")
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

	go func() {
		scannerOut := bufio.NewScanner(p.out)
		for scannerOut.Scan() {
			line := scannerOut.Text()
			parts := strings.SplitN(line, " ", 2)
			if strings.HasPrefix(line, "SH_RES") {
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
			if strings.HasPrefix(line, "SH_RES") {
				p.errChan <- errors.New(parts[1])
				continue
			}
			log.Printf("ERROR LOG %s\n", line)
		}
	}()

	return &p, nil
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

func (p *Phantom) Run(jsFunc string) (string, error) {
	err := p.sendLine("RUN", jsFunc, "END")
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
