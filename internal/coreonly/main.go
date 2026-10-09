// Command coreonly converts one date using only package nepdate. The CI program builds it and
// checks that the binary holds no patro data (ADR-0004 gate 6).
package main

import (
	"os"
	"time"

	nepdate "github.com/sanamhub/nepdate-go"
)

func main() {
	d, err := nepdate.FromGregorian(2024, time.July, 30)
	if err != nil {
		os.Exit(1)
	}
	_, _ = os.Stdout.WriteString(d.String() + "\n")
}
