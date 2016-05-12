package conf

import (
	"strings"
	"sync"
	"testing"
)

type iniStruct struct {
	V1    int8    `ini:"V1"`
	V2    int16   `ini:"V2"`
	V3    int32   `ini:"V3"`
	V4    int64   `ini:"V4"`
	V5    int     `ini:"V5"`
	V6    float32 `ini:"V6"`
	V7    float64 `ini:"V7"`
	V8    string  `ini:"V8"`
	V9    bool    `ini:"V9"`
	mutex *sync.RWMutex
}

func TestIni(t *testing.T) {
	s := `V1 = 100
	V2 = 200
	V3 = 300
	V4 = 400
	V5 = 500
	VV = 999
	#Comments Comments
	V6 = 600.1
	V7 = 700.2
	V8 = hello
	V9 = false
	V10 = ferer
	`

	result := &iniStruct{}
	err := LoadIni(strings.NewReader(s), result)
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	if result.V1 != 100 || result.V2 != 200 || result.V3 != 300 || result.V4 != 400 ||
		result.V5 != 500 || result.V6 != 600.1 || result.V7 != 700.2 || result.V8 != "hello" ||
		result.V9 != false {
		t.Errorf("ini fail")
	}
}
