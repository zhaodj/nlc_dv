package main

import (
	"fmt"
	"io"
	"nlc_dv/marc"
	"os"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func readFile(fp string) {
	f, err := os.Open(fp)
	check(err)
	r := marc.NewReader(f)
	for {
		rc, err := r.Read()
		if err == io.EOF {
			break
		}
		check(err)
		for _, v := range rc.Field {
			fmt.Printf("%d %s\r\n", v.Header, v.Value)
		}
		fmt.Println()
	}
}

func main() {
	file := os.Args[1]
	readFile(file)
}
