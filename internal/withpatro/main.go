// Command withpatro prints the English events of BS 2081-04-15. The CI program builds it to
// show that a binary using package patro does contain the patro names, so the core-only check
// can't pass by accident (ADR-0004 gate 6).
package main

import (
	"os"

	nepdate "github.com/sanamhub/nepdate-go"
	"github.com/sanamhub/nepdate-go/patro"
)

func main() {
	for e := range patro.Of(nepdate.MustNew(2081, nepdate.Shrawan, 15)).Events() {
		_, _ = os.Stdout.WriteString(e.Name(nepdate.English) + "\n")
	}
}
