package piperunner

import "fmt"

var doc string = `# Header

Lorum ipsum ..
`

func Example_single() {
	StartPool()

	resultc := Exec("pandoc -f markdown -t html", []byte(doc))

	result := <-resultc

	if result.err != nil {
		fmt.Printf("error: %v\n", result.err)
	} else {
		fmt.Printf("%v\n", string(result.text))
	}

	// output:
	// <h1 id="header">Header</h1>
	// <p>Lorum ipsum ..</p>
}
