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
	index     = map[int]*IndexItem{}
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

type Document struct {
	Fields []Field
}

type Term struct {
	Field string
	Value string
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
	Search() *IndexItem
}

type TermQuery struct {
	T *Term
}

func (q *TermQuery) Match(t *Term) bool {
	return &t == &q.T
}

func (q *TermQuery) Search() *IndexItem {
	tid, e := lexicon[*q.T]
	if !e {
		return nil
	}
	ii, e := index[tid]
	if !e {
		return nil
	}
	return ii
}

type BooleanQuery struct {
	Q1  Query
	Q2  Query
	Rel Boolean
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

func mergeShould(i1 *IndexItem, i2 *IndexItem) (res *IndexItem) {
	ci1, ci2 := i1, i2
	if i1.docId <= i2.docId {
		res = i1.clone()
		ci1 = ci1.next
	} else {
		res = i2.clone()
		ci2 = ci2.next
	}
	res.next = nil
	cur := res
	for {
		if ci1 == nil {
			if ci2 == nil {
				break
			}
			cur.next = ci2.clone()
			break
		}
		if ci2 == nil {
			cur.next = ci1.clone()
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
	}
	return res
}

func mergeMust(i1 *IndexItem, i2 *IndexItem) (res *IndexItem) {
	cur := res
	ci1, ci2 := i1, i2
	for {
		if ci1.docId == ci2.docId {
			if cur == nil {
				cur = ci1.clone()
				cur.next = nil
				res = cur
			} else {
				cur.next = ci1.clone()
				cur = cur.next
				cur.next = nil
			}
			ci1 = ci1.next
			ci2 = ci2.next
		} else if ci1.docId < ci2.docId {
			ci1 = ci1.next
		} else if ci1.docId > ci2.docId {
			ci2 = ci2.next
		}
		if ci1 == nil || ci2 == nil {
			break
		}
	}
	return res
}

func (q *BooleanQuery) Search() *IndexItem {
	ii1 := q.Q1.Search()
	ii2 := q.Q2.Search()
	if ii1 == nil {
		return ii2
	}
	if ii2 == nil {
		return ii1
	}
	if q.Rel == MUST {
		return mergeMust(ii1, ii2)
	}
	return mergeShould(ii1, ii2)
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
				pre, ex := index[tid]
				if ex {
					pre = pre.add(ii)
				} else {
					pre = ii
				}
				index[tid] = pre
			}
		}
	}
}

func (s *Searcher) Find(q Query) []*Document {
	ii := q.Search()
	if ii == nil {
		return nil
	}
	res := []*Document{}
	for {
		res = append(res, docs[ii.docId])
		if ii.next == nil {
			break
		}
		ii = ii.next
	}
	return res
}
