package marc

import(
    "testing"
    "os"
    "io"
    "fmt"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func TestRead(t *testing.T){
    f, err := os.Open("../中文图书.iso")
	check(err)
	r := NewReader(f)
	for {
		rc, err := r.Read()
		if err == io.EOF {
			break
		}
		check(err)
        fmt.Println(rc.Orig)
	}
}
