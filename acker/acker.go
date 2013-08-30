package main

import (
	"fmt"
	"os"
	"strconv"
)

type anum int32

func acker(n, m anum) anum {
	if n == 0 {
		return m + 1
	} else if m == 0 {
		return acker(n-1, 1)
	} else {
		return acker(n-1, acker(n, m-1))
	}
}

func main() {
	n,e := arg(1)
	m,f := arg(2)
	if e != nil || f != nil {
		os.Exit(1)
	}
	
	a := acker(n, m)
	fmt.Printf("acker(%d,%d) = %d\n", n, m, a)
	fmt.Printf("%#v\n", os.Args)
}

func arg(i int) (int32, error) {
	n,e := strconv.ParseInt(os.Args[i], 0, 32)
	if e != nil {
		fmt.Fprintln(os.Stderr, e.Error())
		fmt.Fprintln(os.Stderr, "Nix verstehen: ", os.Args[i])
	}

	return anum(n), e
}

