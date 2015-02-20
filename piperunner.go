package piperunner

import (
	"errors"
	"io/ioutil"
	"os/exec"
	"sync"
	"time"
)

// Result contains the pandoc output
type Result struct {
	text []byte // result is the pandoc output, could be anything
	err  error
}

type job struct {
	cmd     string      // command to execute
	input   []byte      // input to send over stdin
	resultC chan Result // channel on which to receive the result
}

type Config struct {
	NrWorkers           int           // max number of parallel jobs
	WaitForWorkerMs     time.Duration // mx ms to wait for free worker
	WaitForCompletionMs time.Duration // max ms to wait for completion
}

var (
	jobs   chan job            // queue with jobs waiting
	once   sync.Once           // assure only one pool is started
	config Config    = Config{ // default configuration
		NrWorkers:           3,
		WaitForWorkerMs:     500,
		WaitForCompletionMs: 500,
	}
)

var startPool = func() {
	jobs = make(chan job, 0)

	for i := 0; i < config.NrWorkers; i++ {
		go func() {
			for {
				select {
				case j := <-jobs:
					j.resultC <- startCmd(j.cmd, j.input)
				}
			}
		}()
	}
}

// startPool() starts the pool of workers
func StartPool() {
	once.Do(startPool)
}

// Exec runs command with input send over stdin
func Exec(cmd string, input []byte) <-chan Result {
	j := job{cmd: cmd, input: input, resultC: make(chan Result)}

	go func() {
		select {
		case jobs <- j:
			// sending the job on the channnel that might block is
			// enough for now - if sending does not happen soon enough
			// it will time out

		case <-time.After(config.WaitForWorkerMs * time.Millisecond):
			j.resultC <- Result{text: make([]byte, 0), err: errors.New("wait for worker timed out")}
		}
	}()

	return j.resultC
}

func startCmd(cmd string, in []byte) Result {
	// we have a time out and a job to do.
	complete := make(chan struct{})
	result := Result{text: make([]byte, 0), err: errors.New("wait for completion timed out")}

	go func() {
		// can we rewrite result here?

		close(complete)
	}()

	select {
	case <-complete:
		return result
	case <-time.After(config.WaitForCompletionMs * time.Millisecond):
		return result
	}
}

func execute(cmd string, in []byte) Result {
	command := exec.Command("bash", "-c", cmd)

	stdin, err := command.StdinPipe()

	if err != nil {
		return Result{text: make([]byte, 0), err: err}
	}

	stdout, err := command.StdoutPipe()

	if err != nil {
		return Result{text: make([]byte, 0), err: err}
	}

	if err := command.Start(); err != nil {
		return Result{text: make([]byte, 0), err: err}
	}

	// write in to stdin
	stdin.Write(in)
	stdin.Close()

	// read stdout as a result
	result, err := ioutil.ReadAll(stdout)

	if err != nil {
		return Result{text: make([]byte, 0), err: err}
	}

	return Result{text: result, err: nil}
}
