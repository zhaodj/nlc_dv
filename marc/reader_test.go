package marc

import (
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"testing"
)

func check(e error) {
	if e != nil {
		debug.PrintStack()
		panic(e)
	}
}

func TestRead(t *testing.T) {
	f, err := os.Open("../中文图书_to_sunqian.iso") //../demo.iso 提交-中文图书C-0008
	check(err)
	r := NewReader(f, 0, false)
	for {
		fmt.Println("row")
		rc, err := r.Read()
		if err == io.EOF {
			fmt.Println("eof")
			break
		}
		check(err)
		for _, v := range rc.Field {
			fmt.Println(v.Header, v.Value)
		}
		//fmt.Println(rc.Orig)
	}
}
