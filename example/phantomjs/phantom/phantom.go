package phantom

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

type Phantom struct {
	cmd    *exec.Cmd
	in     io.WriteCloser
	out    io.ReadCloser
	errout io.ReadCloser
}

func Start(args ...string) (*Phantom, error) {
	args = append(args, "test.js")
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
		cmd:    cmd,
		in:     inPipe,
		out:    outPipe,
		errout: errPipe,
	}
	err = cmd.Start()

	if err != nil {
		return nil, err
	}

	return &p, nil
}

func (p *Phantom) Exit() error {
	err := p.Load("phantom.exit()")
	if err != nil {
		return err
	}

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
	scannerOut := bufio.NewScanner(p.out)
	scannerErrorOut := bufio.NewScanner(p.errout)
	resMsg := make(chan string)
	errMsg := make(chan error)
	go func() {
		for scannerOut.Scan() {
			line := scannerOut.Text()
			parts := strings.SplitN(line, " ", 2)
			if strings.HasPrefix(line, "SH_RES") {
				resMsg <- parts[1]
				close(resMsg)
				return
			}
			fmt.Printf("LOG %s\n", line)

		}
	}()
	go func() {
		for scannerErrorOut.Scan() {
			line := scannerErrorOut.Text()
			parts := strings.SplitN(line, " ", 2)
			if strings.HasPrefix(line, "SH_RES") {
				errMsg <- errors.New(parts[1])
				close(errMsg)
				return
			}
			fmt.Printf("LOG %s\n", line)
		}
	}()
	select {
	case text := <-resMsg:
		var res interface{}
		err = json.Unmarshal([]byte(text), &res)
		if err != nil {
			return "", err
		}
		return res.(string), nil
	case err := <-errMsg:
		return "", err
	}

}

func (p *Phantom) Load(jsCode string) error {
	return p.sendLine("EVAL", jsCode, "END")
}

func (p *Phantom) sendLine(lines ...string) error {
	for _, l := range lines {
		_, err := io.WriteString(p.in, l+"\n")
		if err != nil {
			return errors.New("Cannot Send: `" + l + "`")
		}
	}
	return nil
}
