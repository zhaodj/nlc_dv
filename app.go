package main

import (
	"encoding/json"
	"fmt"
	"github.com/codegangsta/negroni"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"nlc_dv/marc"
	"nlc_dv/search"
	"os"
	"reflect"
	"sort"
	"strconv"
)

var ds *DataStore

type Doc struct {
	Id     int
	Year   int      `json:"year"`
	Name   string   `json:"name"`
	Terms  []string `json:"terms"`
	Desc   string   `json:"desc"`
	Author []string `json:"author"`
	URL    string   `json:"url"`
}

type DataStore struct {
	tn           int
	dn           int
	Lexicon      map[string]int
	Docs         map[int]*Doc
	searcher     *search.Searcher
	yearStatData []*YearStat
	yearStatMap  map[int]*YearStat
}

type YearStat struct {
	Year     int          `json:"year,string"`
	Quantity int          `json:"quantity"`
	Keywords []*WordCount `json:"keywords"`
	words    map[string]int
}

type ByYear []*YearStat

func (s ByYear) Len() int {
	return len(s)
}

func (s ByYear) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s ByYear) Less(i, j int) bool {
	return s[i].Year < s[j].Year
}

type WordCount struct {
	Value string `json:"value"`
	Count int    `json:"count"`
}

type ByCount []*WordCount

func (s ByCount) Len() int {
	return len(s)
}

func (s ByCount) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s ByCount) Less(i, j int) bool {
	return s[i].Count > s[j].Count
}

func (y *YearStat) AddWord(word string) {
	c, e := y.words[word]
	if e {
		y.words[word] = c + 1
	} else {
		y.words[word] = 1
	}
}

func (y *YearStat) initKeywords() {
	l := len(y.words)
	y.Keywords = make([]*WordCount, l, l)
	i := 0
	for k, v := range y.words {
		y.Keywords[i] = &WordCount{k, v}
		i++
	}
	sort.Sort(ByCount(y.Keywords))
}

func (d *DataStore) Add(doc *Doc) {
	d.dn++
	d.Docs[d.dn] = doc
	doc.Id = d.dn
	y, e := d.yearStatMap[doc.Year]
	if !e {
		y = &YearStat{Year: doc.Year, Quantity: 1, words: map[string]int{}}
		d.yearStatMap[doc.Year] = y
	} else {
		y.Quantity++
	}
	for _, v := range doc.Terms {
		id, exists := d.Lexicon[v]
		if !exists {
			d.tn++
			id = d.tn
			d.Lexicon[v] = id
		}
		d.searcher.Put(id, d.dn)
		y.AddWord(v)
	}
}

func (d *DataStore) initYearStat() {
	l := len(d.yearStatMap)
	d.yearStatData = make([]*YearStat, l, l)
	i := 0
	for _, v := range d.yearStatMap {
		v.initKeywords()
		d.yearStatData[i] = v
		i++
	}
	sort.Sort(ByYear(d.yearStatData))
}

func (d *DataStore) searchToDoc(sr *search.SearchResult) ([]*Doc, int) {
	if sr == nil || sr.Docs == nil {
		return nil, 0
	}
	res := []*Doc{}
	for _, v := range sr.Docs {
		for _, f := range v.Fields {
			if f.GetName() == "id" {
				res = append(res, d.Docs[f.GetValue().(int)])
				break
			}
		}
	}
	return res, sr.Total
}

func (d *DataStore) Find(term string, year string, start int, limit int) ([]*Doc, int) {
	var q search.Query
	if term == "" && year == "" {
		return nil, 0
	}
	if term == "" && year != "" {
		q = &search.TermPageQuery{search.TermQuery{&search.Term{"year", year}}, start, limit}
	} else if term != "" && year == "" {
		q = &search.TermPageQuery{search.TermQuery{&search.Term{"term", term}}, start, limit}
	} else {
		q = &search.BooleanQuery{
			&search.TermQuery{&search.Term{"term", term}},
			&search.TermQuery{&search.Term{"year", year}},
			search.MUST,
			start,
			limit}
	}
	fmt.Println(reflect.TypeOf(q))
	return d.searchToDoc(d.searcher.Find(q))
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
				//if err != nil || y < 1949 {
				return nil
			}
			//fmt.Println(y)
			doc.Year = y
			i = i | 1
		case 200:
			doc.Name = marc.ParseSubfield(v.Value, 'a')
			//fmt.Println(doc.Name)
			i = i | 2
		case 606:
			//doc.Terms = marc.ParseAllSubfield(v.Value)
			doc.Terms = []string{marc.ParseSubfield(v.Value, 'a')}
			//fmt.Println(v.Value, doc.Terms)
			if len(doc.Terms) == 0 {
				return nil
			}
			i = i | 4
		case 330:
			doc.Desc = marc.ParseSubfield(v.Value, 'a')
			i = i | 8
		case 701:
			if doc.Author == nil {
				doc.Author = []string{}
			}
			au := marc.ParseSubfield(v.Value, 'a')
			if au != "" {
				doc.Author = append(doc.Author, au)
				if i&16 == 0 {
					i = i | 16
				}
			}
		case 856:
			doc.URL = marc.ParseSubfield(v.Value, 'u')
			i = i | 32
		}
	}
	if (i & 7) < 7 {
		fmt.Printf("%d %s %s\r\n", doc.Year, doc.Name, doc.Desc)
		//fmt.Println(doc.Terms)
		return nil
	}
	return doc
}

func docForSearch(doc *Doc) *search.Document {
	fid := &search.IntField{search.BaseField{true, "id"}, doc.Id}
	fyear := &search.IntField{search.BaseField{true, "year"}, doc.Year}
	fterms := &search.StrSliceField{search.BaseField{true, "term"}, doc.Terms}
	fields := []search.Field{fid, fyear, fterms}
	return &search.Document{fields}
}

func readFile(fp string, skip int, chinese bool) *DataStore {
	searcher := search.NewSearcher()
	ds := &DataStore{
		searcher:    searcher,
		Lexicon:     map[string]int{},
		Docs:        map[int]*Doc{},
		yearStatMap: map[int]*YearStat{},
	}
	f, err := os.Open(fp)
	check(err)
	r := marc.NewReader(f, skip, chinese)
	for {
		rc, err := r.Read()
		if err == io.EOF {
			break
		}
		check(err)
		doc := convert(rc)
		if doc != nil {
			ds.Add(doc)
			searcher.Add(docForSearch(doc))
		}
	}
	ds.initYearStat()
	return ds
}

func home(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("views/index.html")
	t.Execute(w, nil)
}

func writeJson(w http.ResponseWriter, d interface{}) {
	b, err := json.Marshal(d)
	if err != nil {
		fmt.Println("json err: ", err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

func limitStatData(data []*YearStat, limit int) []*YearStat {
	res := make([]*YearStat, len(data))
	for i, item := range data {
		if len(item.Keywords) > limit {
			res[i] = &YearStat{item.Year, item.Quantity, item.Keywords[:limit], item.words}
		} else {
			res[i] = item
		}
	}
	return res
}

func yearJson(w http.ResponseWriter, r *http.Request) {
	fmt.Println(len(ds.yearStatData))
	writeJson(w, limitStatData(ds.yearStatData, 100))
}

func findDoc(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	fmt.Println(q)
	data := map[string]interface{}{}
	start := getIntParam(q, "start", 0)
	limit := getIntParam(q, "limit", 50)
	docs, total := ds.Find(q.Get("word"), q.Get("year"), start, limit)
	data["docs"] = docs
	data["total"] = total
	writeJson(w, data)
}

func getIntParam(q url.Values, key string, def int) int {
	str := q.Get(key)
	res := def
	if str != "" {
		res, _ = strconv.Atoi(str)
		if res < 0 {
			res = 0
		}
	}
	return res
}

func main() {
	file := os.Args[1]
	skip := 0
	if len(os.Args) > 2 {
		skip, _ = strconv.Atoi(os.Args[2])
	}
	ds = readFile(file, skip, false)

	mux := http.NewServeMux()
	mux.HandleFunc("/data.json", yearJson)
	mux.HandleFunc("/search.json", findDoc)
	mux.HandleFunc("/", home)

	n := negroni.Classic()
	s := negroni.NewStatic(http.Dir("static"))
	n.Use(s)
	n.UseHandler(mux)
	n.Run(":3000")
}
