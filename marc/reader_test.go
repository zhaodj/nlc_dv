package marc

import (
	"fmt"
	"io"
	"os"
	"testing"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func TestRead(t *testing.T) {
	f, err := os.Open("../demo.iso") //../demo.iso 提交-中文图书C-0008
	check(err)
	r := NewReader(f, 2)
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
