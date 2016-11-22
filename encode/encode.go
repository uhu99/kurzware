/* Dies ist ein Test
*/

package main

import (
	"fmt"
	"os"
	"encoding/json"
)

type Detail struct {
	Number int
}

type Master struct {
	Head string
	Lines []Detail
}

func main() {
	var d = make([]Detail, 10)
	var m, n Master
	var i int

	for i=len(d); i!=0; {
		i -= 0001
		d[i].Number = i
	}
	m.Head = "Hallo Welt"
	m.Lines = d
	fmt.Printf("%+v\n\n", m)

	b,_ := json.Marshal(m)
	os.Stdout.Write(b)
	fmt.Println()

	json.Unmarshal(b, &n)
	fmt.Printf("%+v\n\n", n)	
}
