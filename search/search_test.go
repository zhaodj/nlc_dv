package search

import (
	"testing"
)

func TestEqual(t *testing.T) {
	m := map[Term]int{}
	mp := map[*Term]int{}
	t1 := Term{"a", "a"}
	t2 := Term{"a", "a"}
	m[t1] = 1
	m[t2] = 2
	if len(m) != 1 {
		t.Error("not equal")
	}
	mp[&t1] = 1
	mp[&t2] = 2
	if len(mp) != 2 {
		t.Error("not equal")
	}
}

func TestSearch(t *testing.T){
    idName := "id"
    termsName := "terms"
    doc1 := &Document{
        []Field{
            &IntField{BaseField{true,idName}, 1},&StrSliceField{BaseField{true,termsName},[]string{"中国","北京"}},
        },
    }
    doc2 := &Document{
        []Field{
            &IntField{BaseField{true,idName}, 2},&StrSliceField{BaseField{true,termsName},[]string{"台湾","北京","上海"}},
        },
    }
    searcher := NewSearcher()
    searcher.Add(doc1)
    searcher.Add(doc2)
}
