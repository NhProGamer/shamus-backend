package infra

import (
	"bytes"
	"sync"

	jsoniter "github.com/json-iterator/go"
)

// JSON is a faster drop-in replacement for encoding/json
var JSON = jsoniter.ConfigCompatibleWithStandardLibrary

// bufferPool provides reusable byte buffers for JSON operations
var bufferPool = sync.Pool{
	New: func() interface{} {
		return bytes.NewBuffer(make([]byte, 0, 1024))
	},
}

// GetBuffer retrieves a buffer from the pool
func GetBuffer() *bytes.Buffer {
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	return buf
}

// PutBuffer returns a buffer to the pool
func PutBuffer(buf *bytes.Buffer) {
	if buf.Cap() > 64*1024 { // Don't pool buffers larger than 64KB
		return
	}
	bufferPool.Put(buf)
}

// MarshalJSON is a faster JSON marshal using jsoniter
func MarshalJSON(v interface{}) ([]byte, error) {
	return JSON.Marshal(v)
}

// UnmarshalJSON is a faster JSON unmarshal using jsoniter
func UnmarshalJSON(data []byte, v interface{}) error {
	return JSON.Unmarshal(data, v)
}

// MarshalToBuffer marshals JSON into a pooled buffer
// Caller should call PutBuffer when done with the buffer
func MarshalToBuffer(v interface{}) (*bytes.Buffer, error) {
	buf := GetBuffer()
	encoder := JSON.NewEncoder(buf)
	if err := encoder.Encode(v); err != nil {
		PutBuffer(buf)
		return nil, err
	}
	// Remove trailing newline added by Encode
	if buf.Len() > 0 && buf.Bytes()[buf.Len()-1] == '\n' {
		buf.Truncate(buf.Len() - 1)
	}
	return buf, nil
}
