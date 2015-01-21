package search

type Searcher struct {
	index map[int][]int
}

type Field struct {
	Indexed bool
	Name    string
	Value   interface{}
}

type Document struct {
	id     int
	fields []*Field
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
