package main

import (
	"C"

	"encoding/binary"
	"fmt"
	"time"
	"unsafe"

	"github.com/vmihailenco/msgpack/v5"
)

type EventTime struct {
	time.Time
}

func (tm *EventTime) MarshalMsgpack() ([]byte, error) {
	b := make([]byte, 8)
	binary.BigEndian.PutUint32(b, uint32(tm.Unix()))
	binary.BigEndian.PutUint32(b[4:], uint32(tm.Nanosecond()))
	return b, nil
}

func (tm *EventTime) UnmarshalMsgpack(b []byte) error {
	if len(b) != 8 {
		return fmt.Errorf("invalid data length: got %d, wanted 8", len(b))
	}
	sec := binary.BigEndian.Uint32(b)
	usec := binary.BigEndian.Uint32(b[4:])
	tm.Time = time.Unix(int64(sec), int64(usec))
	return nil
}

func GetBytes(data unsafe.Pointer, length int) []byte {
	return C.GoBytes(data, C.int(length))
}

func msgPackToMap(dec *msgpack.Decoder) (map[string]interface{}, error) {
	_, err := dec.DecodeArrayLen() //skip
	if err != nil {
		return nil, err
	}
	_, err = dec.DecodeInterface() // skip
	if err != nil {
		return nil, err
	}
	value, err := dec.DecodeMap()
	if err != nil {
		return nil, err
	}

	if tStr, ok := value["time"]; ok {
		t, err := time.Parse(time.RFC3339, tStr.(string))
		if err != nil {
			return nil, err
		}
		value["time"] = t
	}

	return value, nil
}
