package search

type Seacher struct {
	index map[int][]int
}

func NewSeacher() *Seacher{
    m := map[int][]int{}
    return &Seacher{m}
}

func (s *Seacher) Put(term int, doc int){
    v, exists := s.index[term]
    if !exists{
        v = []int{doc}
    }else{
        repeat := false
        for _, val := range v{
            if val == term{
                repeat = true
                break
            }
        }
        if !repeat{
            v = append(v, doc)
        }
    }
    s.index[term] = v
}

func (s *Seacher) Get(term int) ([]int, bool){
    v, exists := s.index[term]
    return v, exists
}
