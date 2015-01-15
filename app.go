package main

import (
	"fmt"
	"io"
	"nlc_dv/marc"
	"os"
)

type Doc struct{
    Year int
    Name string
    Terms []string
}

type DataStore struct{
    tn int
    dn int
    Lexicon map[string]int
    Docs map[int]*Doc
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func convert(r *marc.Record) (doc *Doc){
    doc = &Doc{}
    i := 0
    for _, v := range r.Field{
        switch v.Header{
        case 100:
            i++
        case 200:
            i++
        case 606:
            i++
        }
    }
    if i < 3{
        return nil
    }
    return doc
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
