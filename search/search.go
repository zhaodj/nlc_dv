package search

import (
	"strconv"
)

type Boolean int

const (
	MUST   Boolean = iota
	SHOULD Boolean = iota
)

var (
	docCurId  = 0
	termCurId = 0
	docs      = map[int]*Document{}
	lexicon   = map[Term]int{}
	indexes = map[int]*Index{}
)

type Searcher struct {
	index map[int][]int
}

type Field interface {
	IsIndexed() bool
	Terms() []Term
	GetName() string
	GetValue() interface{}
}

type BaseField struct {
	Indexed bool
	Name    string
}

func (f *BaseField) IsIndexed() bool {
	return f.Indexed
}

func (f *BaseField) GetName() string {
	return f.Name
}

type IntField struct {
	BaseField
	Value int
}

func (f *IntField) Terms() []Term {
	return []Term{Term{f.Name, strconv.Itoa(f.Value)}}
}

func (f *IntField) GetValue() interface{} {
	return f.Value
}

type StrSliceField struct {
	BaseField
	Value []string
}

func (f *StrSliceField) Terms() []Term {
	res := []Term{}
	for _, v := range f.Value {
		res = append(res, Term{f.Name, v})
	}
	return res
}

func (f *StrSliceField) GetValue() interface{} {
	return f.Value
}

type SearchResult struct {
	Docs []*Document
	Total int
}

type Document struct {
	Fields []Field
}

type Term struct {
	Field string
	Value string
}

type Index struct{
	Item *IndexItem
	Size int
}

type IndexItem struct {
	docId int
	next  *IndexItem
}

func (i *IndexItem) add(item *IndexItem) *IndexItem {
	if i.docId > item.docId {
		item.next = i
		i.next = nil
		return item
	} else if i.docId < item.docId {
		if i.next != nil {
			i.next = i.next.add(item)
		} else {
			i.next = item
		}
	}
	return i
}

func (i *IndexItem) clone() *IndexItem {
	return &IndexItem{i.docId, i.next}
}

type Query interface {
	Match(t *Term) bool
	Search() *Index
}

type TermQuery struct {
	T *Term
}

func (q *TermQuery) Match(t *Term) bool {
	return &t == &q.T
}

func (q *TermQuery) search() *Index {
	tid, e := lexicon[*q.T]
	if !e {
		return nil
	}
	ii, e := indexes[tid]
	if !e {
		return nil
	}
	return ii
}

func (q *TermQuery) Search() *Index {
	return q.search()
}

type TermPageQuery struct {
	TermQuery
	Start int
	Limit int
}

func (q *TermPageQuery) Search() *Index {
	ii := q.search()
	res := &Index{Size:ii.Size}
	cur := ii.Item
	for i,l:=0,0;;i++{
		if i < q.Start{
			cur = cur.next
			if cur == nil{
				break
			}
		}else {
			if i == q.Start{
				cur = cur.clone()
				res.Item = cur
				l++
			}
			if cur.next == nil{
				break
			}
			if l >=	q.Limit{
				cur.next = nil
				break
			}
			cur.next = cur.next.clone()
			cur = cur.next
			l++
		}
	}
	return res
}

type BooleanQuery struct {
	Q1  Query
	Q2  Query
	Rel Boolean
	Start int
	Limit int
}

func (q *BooleanQuery) Match(t *Term) bool {
	res := false
	switch q.Rel {
	case MUST:
		res = q.Q1.Match(t) && q.Q2.Match(t)
	case SHOULD:
		res = q.Q1.Match(t) || q.Q2.Match(t)
	}
	return res
}

func mergeShould(i1 *Index, i2 *Index, start int, limit int) (res *Index) {
	total := i1.Size + i2.Size
	ci1, ci2 := i1.Item, i2.Item
	var cur *IndexItem
	res = &Index{Size:total}
	i := 0
	if i1.Item.docId <= i2.Item.docId {
		cur = i1.Item.clone()
	} else {
		cur = i2.Item.clone()
	}
	cur.next = nil
	for {
		if start == i{
			res.Item = cur
		}
		if ci1 == nil {
			if ci2 == nil {
				break
			}
			cur.next = ci2.clone()
			i = i + 1
			break
		}
		if ci2 == nil {
			cur.next = ci1.clone()
			i = i + 1
			break
		}
		if ci1.docId == ci2.docId {
			cur.next = ci1.clone()
			cur = cur.next
			ci1 = ci1.next
			ci2 = ci2.next
		} else if ci1.docId > ci2.docId {
			cur.next = ci2.clone()
			cur = cur.next
			ci2 = ci2.next
		} else {
			cur.next = ci1.clone()
			cur = cur.next
			ci1 = ci1.next
		}
		i = i + 1
		if(i >= limit){
			break;
		}
	}
	return res
}

func mergeMust(i1 *Index, i2 *Index, start int, limit int) (res *Index) {
	res = &Index{Size:0}
	var cur *IndexItem
	ci1, ci2 := i1.Item, i2.Item
	var last *IndexItem
	for {
		if ci1.docId == ci2.docId {
			if cur == nil {
				cur = ci1.clone()
				cur.next = nil
			} else {
				cur.next = ci1.clone()
				cur = cur.next
				cur.next = nil
			}
			ci1 = ci1.next
			ci2 = ci2.next
			if start == res.Size{
				res.Item = cur
			}
			res.Size = res.Size + 1
			if limit == res.Size{
				last = cur
			}
		} else if ci1.docId < ci2.docId {
			ci1 = ci1.next
		} else if ci1.docId > ci2.docId {
			ci2 = ci2.next
		}
		if ci1 == nil || ci2 == nil {
			break
		}
	}
	if last != nil{
		last.next = nil
	}
	return res
}

func (q *BooleanQuery) Search() *Index {
	ii1 := q.Q1.Search()
	ii2 := q.Q2.Search()
	if ii1 == nil {
		return ii2
	}
	if ii2 == nil {
		return ii1
	}
	if q.Rel == MUST {
		return mergeMust(ii1, ii2, q.Start, q.Limit)
	}
	return mergeShould(ii1, ii2, q.Start, q.Limit)
}

func NewSearcher() *Searcher {
	m := map[int][]int{}
	return &Searcher{m}
}

func (s *Searcher) Put(term int, doc int) {
	v, exists := s.index[term]
	if !exists {
		v = []int{doc}
	} else {
		repeat := false
		for _, val := range v {
			if val == term {
				repeat = true
				break
			}
		}
		if !repeat {
			v = append(v, doc)
		}
	}
	s.index[term] = v
}

func (s *Searcher) Get(term int) ([]int, bool) {
	v, exists := s.index[term]
	return v, exists
}

func (s *Searcher) Add(doc *Document) {
	id := docCurId
	docs[id] = doc
	docCurId++
	for _, f := range doc.Fields {
		ts := f.Terms()
		if ts != nil {
			for _, t := range ts {
				ii := &IndexItem{docId: id}
				tid, e := lexicon[t]
				if !e {
					tid = termCurId
					termCurId++
					lexicon[t] = tid
				}
				idx, ex := indexes[tid]
				if !ex {
					idx = &Index{ii, 1}
					indexes[tid] = idx
				} else {
					idx.Item.add(ii);
					idx.Size = idx.Size + 1
				}
			}
		}
	}
}

func (s *Searcher) Find(q Query) *SearchResult {
	res := &SearchResult{[]*Document{}, 0}
	i := q.Search()
	if i != nil {
		res.Total = i.Size
		ii := i.Item
		for {
			res.Docs = append(res.Docs, docs[ii.docId])
			if ii.next == nil {
				break
			}
			ii = ii.next
		}
	}
	return res
}
