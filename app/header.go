package main

import (
	"strings"
)

type Header map[string]string

func (h Header) Get(key string) string {
	v, ok := h[strings.ToLower(key)]
	if !ok {
		return ""
	}
	return v
}

func (h Header) Set(key, value string) {
	h[strings.ToLower(key)] = value
}
