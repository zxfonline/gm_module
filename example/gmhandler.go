package main

import (
	"errors"
	"fmt"
	"reflect"

	gm "github.com/zxfonline/gm_module"
	"github.com/zxfonline/json"
)

func main() {
	gm.RegistHander(reflect.ValueOf(gmHandler{}))
	r := gm.HandleCMD(`Hello("true")`)
	bb, err := json.Marshal(r)
	if err != nil {
		panic(err)
	}
	fmt.Printf("getvalue=%s\n", string(bb))
}

type gmHandler struct {
}

func (h gmHandler) Hello(bol bool) (string, error) {
	if bol {
		return "", errors.New("hello gm tool error")
	} else {
		return "hello gm tool", nil
	}
}
