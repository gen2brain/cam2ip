//go:build !amd64
// +build !amd64

package image

import (
	"encoding/base64"
)

func EncodeToString(src []byte) string {
	return base64.StdEncoding.EncodeToString(src)
}
