package marc

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io"
	"io/ioutil"
	"strconv"
)

const (
	maxLen          int   = 99999
	recordSeparator byte  = 29 //1d
	fieldSeparator  byte  = 30 //1e
	subSeparator    byte  = 31 //1f
	RecordSeparator int32 = 29
	FieldSeparator  int32 = 30
	SubSeparator    int32 = 31
)

var (
	ErrMarc = errors.New("MARC read error")
)

type Reader struct {
	line int
	r    *bufio.Reader
	skip int
}

type Record struct {
	Label *RecordLabel
	Dict  []*RecordDict
	Field []*RecordField
	Orig  string
}

type RecordLabel struct {
	Length    int
	dataStart int
}

type RecordDict struct {
	Tag        int
	Length     int
	FieldStart int
}

type RecordField struct {
	Header int
	Value  string
}

func ParseSubfield(field string, start int32) string {
	r := []rune(field)
	i, j := 2, 2
	for {
		if j > i {
			if r[j] == SubSeparator || r[j] == FieldSeparator {
				break
			}
			j++
		} else if r[i] == SubSeparator && r[i+1] == start {
			i += 2
			j = i + 1
		} else {
			i++
			j++
		}
	}
	if j > i {
		return string(r[i:j])
	}
	return ""
}

func ParseAllSubfield(field string) []string {
	res := []string{}
	r := []rune(field)
	i, j := 2, 2
	for {
		if j > i {
			if r[j] == SubSeparator {
				res = append(res, string(r[i:j]))
				i = j
				continue
			} else if r[j] == FieldSeparator {
				res = append(res, string(r[i:j]))
				break
			}
			j++
		} else if r[i] == SubSeparator && r[i+1] >= 48 && r[i+1] <= 122 {
			i += 2
			j = i + 1
		} else {
			i++
			j++
		}
	}
	return res
}

func NewReader(r io.Reader, skip int) *Reader {
	return &Reader{
		r:    bufio.NewReaderSize(r, maxLen),
		skip: skip,
	}
}

func (r *Reader) Read() (record *Record, err error) {
	for {
		record, err = r.parseRecord()
		if record != nil {
			break
		}
		if err != nil {
			return nil, err
		}
	}
	return record, nil
}

func (r *Reader) readLine() (line []byte, err error) {
	line, err = r.r.ReadBytes(recordSeparator)
	if err != nil {
		return nil, err
	}
	if r.skip > 0 {
		_, err = r.r.Discard(r.skip)
		if err != nil {
			return nil, err
		}
	}
	return line, nil
}

func (r *Reader) readRecord() (record *Record, err error) {
	r.line++
	record = &Record{}
	record.Label, err = r.readLabel()
	line := make([]byte, record.Label.Length)
	other := line[24:len(line)]
	ol, err := r.r.Read(other)
	if err != nil {
		return nil, err
	}
	fmt.Println("read other: ", ol)
	/*
		if len(other) != ol {
			return nil, ErrMarc
		}*/
	record.Orig, _ = decode(line)
	if err != nil {
		return nil, err
	}
	ds := 0
	record.Dict, ds, err = r.parseDict(line)
	if err != nil {
		return nil, err
	}
	record.Field, err = r.parseField(record.Dict, line[ds:len(line)])
	if err != nil {
		return nil, err
	}
	return record, nil
}

func (r *Reader) parseRecord() (record *Record, err error) {
	r.line++
	record = &Record{}
	line, err := r.readLine()
	if err != nil {
		return nil, err
	}
	record.Orig, _ = decode(line)
	record.Label, err = r.parseLabel(line)
	if err != nil {
		return nil, err
	}
	ds := 0
	record.Dict, ds, err = r.parseDict(line)
	if err != nil {
		return nil, err
	}
	record.Field, err = r.parseField(record.Dict, line[ds:])
	if err != nil {
		return nil, err
	}
	return record, nil
}

func (r *Reader) parseField(dict []*RecordDict, line []byte) (field []*RecordField, err error) {
	for _, d := range dict {
		f := line[d.FieldStart : d.FieldStart+d.Length]
		s, err := decode(f)
		if err != nil {
			return nil, err
		}
		field = append(field, &RecordField{d.Tag, s})
	}
	return field, nil
}

func decode(field []byte) (s string, err error) {
	i := bytes.NewReader(field)
	o := transform.NewReader(i, simplifiedchinese.GB18030.NewDecoder())
	d, err := ioutil.ReadAll(o)
	if err != nil {
		return "", err
	}
	return string(d), nil
}

func (r *Reader) parseDict(line []byte) (dict []*RecordDict, ds int, err error) {
	i := 24
	for {
		if line[i] == fieldSeparator {
			i++
			break
		}
		t, l, s := string(line[i:i+3]), string(line[i+3:i+7]), string(line[i+7:i+12])
		i += 12
		ti, err := strconv.Atoi(t)
		if err != nil {
			return nil, i, err
		}
		li, err := strconv.Atoi(l)
		if err != nil {
			return nil, i, err
		}
		si, err := strconv.Atoi(s)
		if err != nil {
			return nil, i, err
		}
		dict = append(dict, &RecordDict{ti, li, si})
	}
	return dict, i, nil
}

func (r *Reader) readLabel() (label *RecordLabel, err error) {
	lb := make([]byte, 24)
	l, err := r.r.Read(lb)
	if err != nil {
		return nil, err
	}
	fmt.Println("read label: ", l)
	/*
		if l != 24 {
			return nil, ErrMarc
		}*/
	return r.parseLabel(lb)
}

func (r *Reader) parseLabel(line []byte) (label *RecordLabel, err error) {
	label = &RecordLabel{}
	label.Length, err = strconv.Atoi(string(line[:5]))
	if err != nil {
		return nil, err
	}
	//fmt.Println("record length: ", label.Length)
	label.dataStart, err = strconv.Atoi(string(line[12:17]))
	if err != nil {
		return nil, err
	}
	return label, err
}
