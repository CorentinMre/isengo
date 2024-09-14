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
