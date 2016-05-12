package conf

import (
	"sync"
	"testing"
)

import "strings"

type tt struct {
	V1 int8
	V2 int16
	V3 int32
	V4 int64
	V5 int
	V6 float32
	V7 float64
	V8 string
	V9 bool
}

type csvStruct struct {
	Values []tt
	mutex  *sync.RWMutex
}

func TestCsv(t *testing.T) {
	s := `1,  2,   3,  4,  5,   6,   7,         8,           9
	    12, 123, 88, 39, 343, 89.1, 98.2312,  hello world, true
		22, 125, 99, 21, 332, 23.9, 1231.22,  chen,     false
		
		`

	var results csvStruct
	err := LoadCsv(strings.NewReader(s), &results, true)
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	if len(results.Values) != 2 {
		t.Errorf("slice size = %d", len(results.Values))
		return
	}

	pt := results.Values[0]
	if pt.V1 != 12 || pt.V2 != 123 || pt.V3 != 88 || pt.V4 != 39 || pt.V5 != 343 ||
		pt.V6 != 89.1 || pt.V7 != 98.2312 || pt.V8 != "hello world" || pt.V9 != true {
		t.Errorf("fail")
	}

	pt = results.Values[1]
	if pt.V1 != 22 || pt.V2 != 125 || pt.V3 != 99 || pt.V4 != 21 || pt.V5 != 332 ||
		pt.V6 != 23.9 || pt.V7 != 1231.22 || pt.V8 != "chen" || pt.V9 != false {
		t.Errorf("fail")
	}
}
