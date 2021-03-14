package tesla

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"io"
	"strconv"
	"strings"
)

func userAgent() string {
	const prefixBytes = 10
	var buf [prefixBytes + 8]byte
	if _, err := io.ReadFull(rand.Reader, buf[:]); err != nil {
		// I know.
		panic(err)
	}

	var b strings.Builder
	b.WriteString(hex.EncodeToString(buf[:prefixBytes]))
	b.WriteRune('/')
	b.Write(strconv.AppendUint(nil, binary.BigEndian.Uint64(buf[prefixBytes:]), 10))
	return b.String()
}
