package client

import "strconv"

type Response struct {
	Event   string
	Project string
	Key     []byte
	Value   []byte
}

func (r *Response) String(def string) string {
	if len(r.Value) == 0 {
		return def
	}

	return string(r.Value)
}

func (r *Response) Int(def int) int {
	if len(r.Value) == 0 {
		return def
	}

	val, err := strconv.Atoi(string(r.Value))
	if err != nil {
		return 0
	}

	return val
}

func (r *Response) Float64(def float64) float64 {
	if len(r.Value) == 0 {
		return def
	}

	val, err := strconv.ParseFloat(string(r.Value), 64)
	if err != nil {
		return 0
	}

	return val
}

func (r *Response) Bytes(def []byte) []byte {
	if len(r.Value) == 0 {
		return def
	}

	return r.Value
}
