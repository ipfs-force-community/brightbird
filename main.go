package main

import (
	"fmt"
	"reflect"

	"github.com/fluent/fluent-bit-go/output"
	"github.com/ugorji/go/codec"
)

func main() {
	dec.handle = new(codec.MsgpackHandle)
	dec.handle.SetBytesExt(reflect.TypeOf(output.FLBTime{}), 0, &output.FLBTime{})

	b = C.GoBytes(data, C.int(length))

	fmt.Println("data: ", b)
	fmt.Println("dataend")
	dec.mpdec = codec.NewDecoderBytes(b, dec.handle)

}
