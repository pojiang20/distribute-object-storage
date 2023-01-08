package utils

import (
	"log"
	"net/http"
	"strconv"
	"strings"
)

const RANGE_FIELD_LENGTH = len("bytes=")

func GetObjectName(url string) string {
	url = strings.TrimSpace(url)
	components := strings.Split(url, "/")
	return components[len(components)-1]
}

func GetHashFromHeader(h http.Header) string {
	digest := h.Get("digest")
	if len(digest) < 9 {
		return ""
	}
	if digest[:8] != "SHA-256=" {
		return ""
	}
	return digest[8:]
}

func GetSizeFromHeader(h http.Header) int64 {
	size, _ := strconv.ParseInt(h.Get("content-length"), 0, 64)
	return size
}

func GetOffsetFromHeader(h http.Header) (int64, int64) {
	byteRange := h.Get("range")
	log.Printf("range: %s\n", byteRange)
	if len(byteRange) < RANGE_FIELD_LENGTH {
		return 0, 0
	}
	if byteRange[:RANGE_FIELD_LENGTH] != "bytes=" {
		return 0, 0
	}
	bytesPositions := strings.Split(byteRange[RANGE_FIELD_LENGTH:], "-")
	log.Println(bytesPositions)
	offset, _ := strconv.ParseInt(bytesPositions[0], 0, 64)
	end, _ := strconv.ParseInt(bytesPositions[1], 0, 64)
	log.Println("offset[%d]-end[%d]\n", offset, end)

	return offset, end
}
