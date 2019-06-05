package tile

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strings"
)

func str2hex(text string) string {

	hasher := md5.New()
	hasher.Write([]byte(text))

	enc := hex.EncodeToString(hasher.Sum(nil))
	code := enc[0:6]

	return fmt.Sprintf("#%s", code)
}

func escapeXMLString(text string) string {

	// there must be a built-in function to do this...
	// (20190529/thisisaaronland)

	text = strings.Replace(text, "&", "&amp;", -1)
	text = strings.Replace(text, ">", "&gt;", -1)
	text = strings.Replace(text, "<", "&lt;", -1)

	return text
}
