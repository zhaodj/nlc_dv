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