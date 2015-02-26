package piperunner

import (
	"fmt"
	"strconv"
	"sync"
)

var doc string = `# Header

Lorum ipsum ..
`

func Example_single() {
	StartPool()

	resultc := Exec("pandoc -f markdown -t html", []byte(doc))

	result := <-resultc

	if result.Err != nil {
		fmt.Printf("error: %v\n", result.Err)
	} else {
		fmt.Printf("%v\n", string(result.Text))
	}

	// output:
	// <h1 id="header">Header</h1>
	// <p>Lorum ipsum ..</p>
}

func Example_several() {
	StartPool()

	// a bit contrived example of n independent pandoc jobs
	n := 50

	cc := make([]<-chan Result, n)
	rr := make([]Result, n)

	for i := 0; i < n; i++ {
		cc[i] = Exec("pandoc -f html -t markdown",
			[]byte("<p>Lorum ipsum "+strconv.Itoa(i)+"</p>"))
	}

	var wg sync.WaitGroup

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(j int) {
			rr[j] = <-cc[j]
			wg.Done()
		}(i)
	}

	wg.Wait()

	for i := 0; i < n; i++ {
		if rr[i].Err != nil {
			fmt.Printf("error: %v\n", rr[i].Err)
		} else {
			fmt.Printf("%v", string(rr[i].Text))
		}
	}

	// output:
	// Lorum ipsum 0
	// Lorum ipsum 1
	// Lorum ipsum 2
	// Lorum ipsum 3
	// Lorum ipsum 4
	// Lorum ipsum 5
	// Lorum ipsum 6
	// Lorum ipsum 7
	// Lorum ipsum 8
	// Lorum ipsum 9
	// Lorum ipsum 10
	// Lorum ipsum 11
	// Lorum ipsum 12
	// Lorum ipsum 13
	// Lorum ipsum 14
	// Lorum ipsum 15
	// Lorum ipsum 16
	// Lorum ipsum 17
	// Lorum ipsum 18
	// Lorum ipsum 19
	// Lorum ipsum 20
	// Lorum ipsum 21
	// Lorum ipsum 22
	// Lorum ipsum 23
	// Lorum ipsum 24
	// Lorum ipsum 25
	// Lorum ipsum 26
	// Lorum ipsum 27
	// Lorum ipsum 28
	// Lorum ipsum 29
	// Lorum ipsum 30
	// Lorum ipsum 31
	// Lorum ipsum 32
	// Lorum ipsum 33
	// Lorum ipsum 34
	// Lorum ipsum 35
	// Lorum ipsum 36
	// Lorum ipsum 37
	// Lorum ipsum 38
	// Lorum ipsum 39
	// Lorum ipsum 40
	// Lorum ipsum 41
	// Lorum ipsum 42
	// Lorum ipsum 43
	// Lorum ipsum 44
	// Lorum ipsum 45
	// Lorum ipsum 46
	// Lorum ipsum 47
	// Lorum ipsum 48
	// Lorum ipsum 49
}

func Example_stderr() {
	StartPool()

	// pandoc will write to stderr about unknown writer noSuchWriter
	resultc := Exec("pandoc -f markdown -t noSuchWriter", []byte(doc))

	result := <-resultc

	if result.Err != nil {
		fmt.Printf("error: %q - %q\n", result.Text, result.Err)
	} else {
		fmt.Printf("%v\n", string(result.Text))
	}

	// output:
	// error: "pandoc: Unknown writer: nosuchwriter\n" - "stderr"
}
