package gobash

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

func Exec(name string, args ...string) ([]byte, error) {
	cmdName, err := exec.LookPath(name)
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(cmdName, args...)
	return getResult(cmd)
}

type Result struct {
	StdOut chan string
	Err    error
}

func Run(ctx context.Context, name string, args ...string) *Result {
	result := &Result{StdOut: make(chan string), Err: error(nil)}

	go func() {
		defer func() { close(result.StdOut) }()
		cmdName, err := exec.LookPath(name)
		if err != nil {
			result.Err = err
			return
		}
		cmd := exec.CommandContext(ctx, cmdName, args...)
		handleExec(ctx, cmd, result)
	}()

	return result
}

func handleExec(ctx context.Context, cmd *exec.Cmd, result *Result) {
	result.StdOut <- strings.Join(cmd.Args, " ") + "\n"

	stdout, stderr, err := getCmdReader(cmd)
	if err != nil {
		result.Err = err
		return
	}

	reader := bufio.NewReader(stdout)

	line := ""
	for {
		line, err = reader.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			result.Err = err
			break
		}
		select {
		case result.StdOut <- line:
		case <-ctx.Done():
			result.Err = fmt.Errorf("%v", ctx.Err())
			return
		}
	}

	bytesErr, err := io.ReadAll(stderr)
	if err != nil {
		result.Err = err
		return
	}

	err = cmd.Wait()
	if err != nil {
		if len(bytesErr) != 0 {
			result.Err = errors.New(string(bytesErr))
			return
		}
		result.Err = err
	}
}

func getCmdReader(cmd *exec.Cmd) (stdout io.ReadCloser, stderr io.ReadCloser, err error) {
	stdout, err = cmd.StdoutPipe()
	if err != nil {
		return nil, nil, err
	}

	stderr, err = cmd.StderrPipe()
	if err != nil {
		return nil, nil, err
	}

	err = cmd.Start()
	if err != nil {
		return nil, nil, err
	}

	return stdout, stderr, nil
}

func getResult(cmd *exec.Cmd) ([]byte, error) {
	stdout, stderr, err := getCmdReader(cmd)
	if err != nil {
		return nil, err
	}

	bytes, err := io.ReadAll(stdout)
	if err != nil {
		return nil, err
	}

	bytesErr, err := io.ReadAll(stderr)
	if err != nil {
		return nil, err
	}

	err = cmd.Wait()
	if err != nil {
		if len(bytesErr) != 0 {
			return nil, errors.New(string(bytesErr))
		}
		return nil, err
	}

	return bytes, nil
}
