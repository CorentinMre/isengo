<br>
<p align="center"><img width="400" alt="Logo" src="https://raw.githubusercontent.com/CorentinMre/isengo/main/images/icon.jpg"></a></p>

<br/>

<h2 style="font-family: sans-serif; font-weight: normal;" align="center"><strong>An API for ISEN-OUEST</strong></h2>

<br/>

<h2 style="font-family: sans-serif; font-weight: normal;" align="center"><strong>⚠️ Unofficial !!</strong></h2>

## Description

A [GO](https://go.dev/) API wrapper for ISEN-OUEST, with webAurion information like planning, grades, absences and user info

## Dependencies

- [goquery](https://github.com/PuerkitoBio/goquery)

## Usage
- `go mod init <name-of-your-project>`
- `go get github.com/CorentinMre/isengo/webaurion`

Here is an example script:

```go
package main

import (
	"fmt"
	"github.com/CorentinMre/isengo/webaurion"
)

func main() {
	w := webaurion.NewWebAurion()

	// login
	success, err := w.Login("<username>", "<password>")
	if err != nil || !success {
		fmt.Println("Login failed:", err)
		return
	}

	userInfo, err := w.UserInfo()
	if err != nil {
		fmt.Println("Failed to get user info:", err)
	} else {
		fmt.Printf("User info: %+v\n", userInfo)

		fmt.Println("User info JSON: ", userInfo.JSON())
	}
}

```

## Example for get your grades

```go

...


grades, err := w.GetGrades()
if err != nil {
    fmt.Println("Failed to get grades:", err)
} else {

    fmt.Println("Grades: ", grades.JSON())
}

```

## Example for get your absences

```go

...


absences, err := w.GetAbsences()
if err != nil {
    fmt.Println("Failed to get absences:", err)
} else {
    fmt.Println("Absences: ", absences.JSON())
}

```

## Example for get your planning

```go

...

planning, err := w.GetPlanning()
if err != nil {
    fmt.Println("Failed to get planning:", err)
} else {
    fmt.Println("Planning: ", planning.JSON())
}

```

## Example for get catalog entries

```go

...

// load catalogs
err = w.LoadCatalogs()
if err != nil {
    fmt.Println("Failed to load catalogs:", err)
    return
}

// list catalogs
catalogs := w.ListCatalogs()
for i, cat := range catalogs {
    fmt.Printf("%d. %s\n", i, cat.Name)
}
// =
//0. Catalogue des stages associatifs
//1. Catalogue des stages ouvriers
//2. Catalogue des stages techniciens
//3. Catalogue des stages M1
//4. Catalogue des stages M2
//5. Catalogue des apprentissages

// Get entries from a specific catalog (e.g., index 0) ()
report, err := w.GetCatalogEntries(0) // 0 = Catalogue des stages associatifs
if err != nil {
    fmt.Println("Failed to get catalog entries:", err)
} else {
    totalEntries, _ := report.Get("totalEntries")
    fmt.Printf("Total entries: %d\n", totalEntries)

    // print JSON
    fmt.Println("Catalog entries: ", report.JSON()) // warning: much info
}

// get more info about 1 row (btn "Consulter" on webaurion)
entries, _ := report.Get("entries")
entriesList := entries.([]catalog.CatalogEntry)
if len(entriesList) > 0 {
    details, err := w.GetCatalogEntryDetails(entriesList[0]) // first row details (= button "Consulter" on webaurion)
    if err != nil {
        fmt.Println("Failed to get entry details:", err)
    } else {
        fmt.Println("Entry details: ", details.JSON())
    }
}

```

## LICENSE

Copyright (c) 2022-2024 CorentinMre

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
