package piperunner

import (
	"errors"
	"io/ioutil"
	"os/exec"
	"sync"
	"time"
)

// Result contains the pandoc output or an error
type Result struct {
	Text []byte // result is the pandoc output, could be anything
	Err  error
}

type job struct {
	cmd     string      // command to execute
	input   []byte      // input to send over stdin
	resultC chan Result // channel on which to receive the result
}

// configuration allows for a different configuration
type configuration struct {
	NrWorkers           int           // max number of parallel jobs
	WaitForWorkerMs     time.Duration // mx ms to wait for free worker
	WaitForCompletionMs time.Duration // max ms to wait for completion
}

var (
	jobs   chan job                       // queue with jobs waiting
	once   sync.Once                      // assure only one pool is started
	config configuration = configuration{ // default configuration
		NrWorkers:           3,
		WaitForWorkerMs:     500,
		WaitForCompletionMs: 500,
	}
)

// Config is used to set number of workers, timeout to wait for free
// worker and timeout to wait for worker completion.
// Run Config before StartPool
func Config(nrWorkers int, waitForWorker, waitForCompletion time.Duration) {
	config = configuration{nrWorkers, waitForWorker, waitForCompletion}
}

// startPool() starts the pool of workers, once
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
			j.resultC <- Result{Text: make([]byte, 0), Err: errors.New("wait for worker timed out")}
		}
	}()

	return j.resultC
}

// startPool defines the function that will start the pool
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

// startCmd starts the command and checks for time out
func startCmd(cmd string, in []byte) Result {
	// we have a time out and a job to do.
	complete := make(chan struct{})
	result := Result{Text: make([]byte, 0), Err: errors.New("wait for completion timed out")}

	go func() {
		result = execute(cmd, in)

		close(complete)
	}()

	select {
	case <-complete:
		return result
	case <-time.After(config.WaitForCompletionMs * time.Millisecond):
		return result
	}
}

// execute runs bash, feeds on stdin and reads from stdout
func execute(cmd string, in []byte) Result {
	command := exec.Command("bash", "-c", cmd)

	stdin, err := command.StdinPipe()

	if err != nil {
		return Result{Text: make([]byte, 0), Err: err}
	}

	stdout, err := command.StdoutPipe()

	if err != nil {
		return Result{Text: make([]byte, 0), Err: err}
	}

	stderr, err := command.StderrPipe()

	if err != nil {
		return Result{Text: make([]byte, 0), Err: err}
	}

	// run the command

	if err := command.Start(); err != nil {
		return Result{Text: make([]byte, 0), Err: err}
	}

	defer func() {
		// releases the resources
		command.Wait()
	}()

	// write in to stdin
	stdin.Write(in)
	stdin.Close()

	// read stdout as a result
	result, err := ioutil.ReadAll(stdout)
	stdout.Close()

	if err != nil {
		return Result{Text: make([]byte, 0), Err: err}
	}

	// read stderr for possible errors
	msgs, err := ioutil.ReadAll(stderr)
	stderr.Close()

	if len(msgs) > 0 {
		return Result{Text: msgs, Err: errors.New("stderr")}
	}

	if err != nil {
		return Result{Text: make([]byte, 0), Err: err}
	}

	return Result{Text: result, Err: nil}
}
