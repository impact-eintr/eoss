package utils

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

func GetOffsetFromHeader(h http.Header) int64 {
	byteRange := h.Get("range")
	if len(byteRange) < 7 {
		return 0
	}
	if byteRange[:6] != "bytes=" {
		return 0
	}
	bytePos := strings.Split(byteRange[6:], "-")
	offset, _ := strconv.ParseInt(bytePos[0], 0, 64)
	return offset
}

func GetHashFromHeader(h http.Header) string {
	var digest string
	if h.Get("digest") != "" {
		digest = h.Get("digest")
	} else if h.Get("Digest") != "" {
		digest = h.Get("Digest")
	} else {
		return ""
	}

	if len(digest) < 9 {
		return ""
	}
	if digest[:8] != "SHA-256=" {
		fmt.Println(digest[:8])
		return ""
	}
	return digest[8:]
}

func GetSizeFromHeader(h http.Header) int64 {
	size, _ := strconv.ParseInt(h.Get("content-length"), 0, 64)
	return size
}

func CalculateHash(r io.Reader) string {
	h := sha256.New()
	io.Copy(h, r)
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
