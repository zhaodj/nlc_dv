package main

import (
	"fmt"
	"io"
	"nlc_dv/marc"
	"nlc_dv/search"
	"os"
	"strconv"
    "github.com/codegangsta/negroni"
    "net/http"
    "html/template"
    "encoding/json"
    "sort"
)

var ds *DataStore

type Doc struct {
    Year  int   `json:"year"`
    Name  string    `json:"name"`
    Terms []string  `json:"terms"`
    Desc string `json:"desc"`
    Author string `json:"author"`
}

type DataStore struct {
	tn       int
	dn       int
	Lexicon  map[string]int
	Docs     map[int]*Doc
	searcher *search.Searcher
    yearStatData []*YearStat
    yearStatMap map[int]*YearStat
}

type YearStat struct{
    Year int `json:"year,string"`
    Quantity int `json:"quantity"`
    Keywords []*WordCount `json:"keywords"`
    words map[string]int
}

type ByYear []*YearStat

func (s ByYear) Len() int {
    return len(s)
}

func (s ByYear) Swap(i, j int){
    s[i], s[j] = s[j], s[i]
}

func (s ByYear) Less(i, j int) bool {
    return s[i].Year < s[j].Year
}

type WordCount struct{
    Value string `json:"value"`
    Count int `json:"count"`
}

type ByCount []*WordCount

func (s ByCount) Len() int {
    return len(s)
}

func (s ByCount) Swap(i, j int){
    s[i], s[j] = s[j], s[i]
}

func (s ByCount) Less(i, j int) bool {
    return s[i].Count > s[j].Count
}

func (y *YearStat) AddWord(word string){
    c, e := y.words[word]
    if e{
        y.words[word] = c + 1
    }else{
        y.words[word] = 1
    }
}

func (y *YearStat) initKeywords(){
    l := len(y.words)
    y.Keywords = make([]*WordCount, l, l)
    i := 0
    for k, v := range y.words{
        y.Keywords[i] = &WordCount{k, v}
        i++
    }
    sort.Sort(ByCount(y.Keywords))
}

func (d *DataStore) Add(doc *Doc) {
	d.dn++
	d.Docs[d.dn] = doc
    y, e := d.yearStatMap[doc.Year]
    if !e{
        y = &YearStat{Year:doc.Year,Quantity:1,words:map[string]int{}}
        d.yearStatMap[doc.Year] = y
    }else{
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

func (d *DataStore) initYearStat(){
    l := len(d.yearStatMap)
    d.yearStatData = make([]*YearStat, l, l)
    i := 0
    for _, v := range d.yearStatMap{
        v.initKeywords()
        d.yearStatData[i] = v
        i++
    }
    sort.Sort(ByYear(d.yearStatData))
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
			if err != nil || y < 1949{
				return nil
			}
			//fmt.Println(y)
			doc.Year = y
			i = i|1
		case 200:
			doc.Name = marc.ParseSubfield(v.Value, 'a')
			//fmt.Println(doc.Name)
			i = i|2
		case 606:
			doc.Terms = marc.ParseAllSubfield(v.Value)
			//fmt.Println(doc.Terms)
			if len(doc.Terms) == 0 {
				return nil
			}
			i = i|4
        case 330:
            doc.Desc = marc.ParseSubfield(v.Value, 'a')
            i = i|8
        case 701:
            if i & 16 == 0{
                doc.Author = marc.ParseSubfield(v.Value, 'a')
                i = i|16
            }
		}
	}
	if (i & 7) < 7 {
        fmt.Printf("%d %s %s\r\n",doc.Year,doc.Name,doc.Desc)
        fmt.Println(doc.Terms)
		return nil
	}
	return doc
}

func readFile(fp string) *DataStore {
	searcher := search.NewSearcher()
	ds := &DataStore{
		searcher: searcher,
		Lexicon:  map[string]int{},
		Docs:     map[int]*Doc{},
        yearStatMap: map[int]*YearStat{},
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
    ds.initYearStat()
	return ds
}

func home(w http.ResponseWriter, r *http.Request){
    t, _ := template.ParseFiles("views/index.html")
    t.Execute(w, nil)
}

func writeJson(w http.ResponseWriter, d interface{}){
    b, err := json.Marshal(d)
    if err != nil{
        fmt.Println("json err: ", err)
    }
    w.Header().Set("Content-Type", "application/json")
    w.Write(b)
}

func yearJson(w http.ResponseWriter, r *http.Request){
    fmt.Println(len(ds.yearStatData))
    writeJson(w, ds.yearStatData)
}

func findDoc(w http.ResponseWriter, r *http.Request){
    q := r.URL.Query()
    writeJson(w, ds.Find(q.Get("word")))
}



func main() {
    file := os.Args[1]
    ds = readFile(file)

    mux := http.NewServeMux()
    mux.HandleFunc("/data.json", yearJson)
    mux.HandleFunc("/search.json",findDoc)
    mux.HandleFunc("/", home)

    n := negroni.Classic()
    s := negroni.NewStatic(http.Dir("static"))
    n.Use(s)
    n.UseHandler(mux)
    n.Run(":3000")
}
