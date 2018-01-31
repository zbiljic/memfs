package cmd

import (
	"fmt"
	"log"
	"strconv"
	"time"
)

const (
	maxUint32 = ^uint32(0)
)

func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("TRACE '%s' took %s", name, elapsed)
}

func valueOrEmptyString(slice []string, index int) string {
	if len(slice) > index {
		return slice[index]
	}
	return ""
}

func parseOctalInt(s string) (i int, err error) {
	var tmp int64
	tmp, err = strconv.ParseInt(s, 8, 32)
	if err != nil {
		err = fmt.Errorf("Parsing as octal: %v", err)
		return
	}

	i = int(tmp)
	return
}
