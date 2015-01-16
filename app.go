package main

import (
	"fmt"
	"io"
	"nlc_dv/marc"
	"nlc_dv/search"
	"os"
	"strconv"
)

type Doc struct {
	Year  int
	Name  string
	Terms []string
}

type DataStore struct {
	tn       int
	dn       int
	Lexicon  map[string]int
	Docs     map[int]*Doc
	searcher *search.Searcher
}

func (d *DataStore) Add(doc *Doc) {
	d.dn++
	d.Docs[d.dn] = doc
	for _, v := range doc.Terms {
		id, exists := d.Lexicon[v]
		if !exists {
			d.tn++
			id = d.tn
			d.Lexicon[v] = id
		}
		d.searcher.Put(id, d.dn)
	}
}

func (d *DataStore) Find(term string) []*Doc {
	id, exists := d.Lexicon[term]
	if !exists {
		return nil
	}
	arr, e := d.searcher.Get(id)
	if !e {
		return nil
	}
	res := []*Doc{}
	for _, v := range arr {
		res = append(res, d.Docs[v])
	}
	return res
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func parseYear(f string) (int, error) {
	sf := marc.ParseSubfield(f, 'a')
	r := []rune(sf)
	ys := string(r[9:13])
	return strconv.Atoi(ys)
}

func convert(r *marc.Record) (doc *Doc) {
	doc = &Doc{}
	i := 0
	for _, v := range r.Field {
		switch v.Header {
		case 100:
			y, err := parseYear(v.Value)
			if err != nil {
				return nil
			}
			fmt.Println(y)
			doc.Year = y
			i++
		case 200:
			doc.Name = marc.ParseSubfield(v.Value, 'a')
			fmt.Println(doc.Name)
			i++
		case 606:
			doc.Terms = marc.ParseAllSubfield(v.Value)
			fmt.Println(doc.Terms)
			if len(doc.Terms) == 0 {
				return nil
			}
			i++
		}
	}
	if i < 3 {
		return nil
	}
	fmt.Println()
	return doc
}

func readFile(fp string) *DataStore {
	searcher := search.NewSearcher()
	ds := &DataStore{
		searcher: searcher,
		Lexicon:  map[string]int{},
		Docs:     map[int]*Doc{},
	}
	f, err := os.Open(fp)
	check(err)
	r := marc.NewReader(f)
	for {
		rc, err := r.Read()
		if err == io.EOF {
			break
		}
		check(err)
		doc := convert(rc)
		if doc != nil {
			ds.Add(doc)
		}
	}
	return ds
}

func main() {
	file := os.Args[1]
	ds := readFile(file)
	darr := ds.Find("中国aaa")
	if darr != nil {
		for _, d := range darr {
			fmt.Println(d.Name)
		}
	}
}
