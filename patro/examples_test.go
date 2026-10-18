package patro_test

import (
	"fmt"
	"log"

	nepdate "github.com/sanamhub/nepdate-go"
	"github.com/sanamhub/nepdate-go/patro"
)

func ExampleCoverage() {
	r := patro.Coverage()
	fmt.Println(r.Start(), r.End(), r.Len())
	// Output: 2077/01/01 2084/08/20 2793
}

func ExamplePublicHolidays() {
	n := 0
	for d := range patro.PublicHolidays() {
		if n < 3 {
			fmt.Println(d)
		}
		n++
	}
	fmt.Println("total", n)
	// Output:
	// 2077/01/01
	// 2077/01/19
	// 2077/01/25
	// total 294
}

func ExampleInfo() {
	info := patro.Of(nepdate.MustNew(2081, nepdate.Shrawan, 15))
	t, ok := info.Tithi()
	fmt.Println(t.Name(nepdate.English), ok, info.IsPublicHoliday(), info.EventCount())
	// Output: Dashami true false 3
}

func ExampleInfo_Tithi() {
	t, ok := patro.Of(nepdate.MustNew(2081, nepdate.Shrawan, 15)).Tithi()
	fmt.Println(ok, t.Name(nepdate.Nepali))
	// Output: true दशमी
}

func ExampleInfo_IsPublicHoliday() {
	info := patro.Of(nepdate.MustNew(2081, nepdate.Baishakh, 1))
	fmt.Println(info.IsPublicHoliday())
	// Output: true
}

func ExampleInfo_Events() {
	info := patro.Of(nepdate.MustNew(2081, nepdate.Shrawan, 15))
	for e := range info.Events() {
		fmt.Println(e.Name(nepdate.English))
	}
	// Output:
	// Kheer Khane Din
	// World Day Against Human Trafficking
	// International Friendship Day
}

func ExampleInfo_EventCount() {
	info := patro.Of(nepdate.MustNew(2081, nepdate.Shrawan, 15))
	fmt.Println(info.EventCount())
	// Output: 3
}

func ExampleInfo_Event() {
	info := patro.Of(nepdate.MustNew(2081, nepdate.Shrawan, 15))
	fmt.Println(info.Event(1).Name(nepdate.English))
	// Output: World Day Against Human Trafficking
}

func ExampleTithi() {
	t, ok := patro.Of(nepdate.MustNew(2081, nepdate.Shrawan, 15)).Tithi()
	if !ok {
		log.Fatal("2081-04-15 has tithi data")
	}
	fmt.Println(t.Name(nepdate.English))
	fmt.Println(t.Name(nepdate.Nepali))
	// Output:
	// Dashami
	// दशमी
}

func ExampleEvent() {
	e := patro.Of(nepdate.MustNew(2081, nepdate.Shrawan, 15)).Event(0)
	fmt.Println(e.Name(nepdate.English))
	fmt.Println(e.Name(nepdate.Nepali))
	// Output:
	// Kheer Khane Din
	// खीर खाने दिन
}
