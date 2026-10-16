package patro_test

import (
	"fmt"

	nepdate "github.com/sanamhub/nepdate-go"
	"github.com/sanamhub/nepdate-go/patro"
)

func ExampleOf() {
	d, _ := nepdate.New(2081, nepdate.Baishakh, 1)
	i := patro.Of(d)
	t, _ := i.Tithi()
	fmt.Println(t.Name(nepdate.Nepali), i.IsPublicHoliday())
	for n := 0; n < i.EventCount(); n++ {
		fmt.Println(i.Event(n).Name(nepdate.English))
	}
	// Output:
	// पञ्चमी true
	// New Year
	// Mesh Sankranti
	// Bisket Jatra
}
