# confusables

[![GoDoc](https://godoc.org/github.com/eskriett/confusables?status.svg)](https://godoc.org/github.com/eskriett/confusables)
[![Build Status](https://travis-ci.org/eskriett/confusables.svg?branch=master)](https://travis-ci.org/eskriett/confusables)
[![Go Report Card](https://goreportcard.com/badge/github.com/eskriett/confusables)](https://goreportcard.com/report/github.com/eskriett/confusables)

Unicode confusable detection

## Overview

```go
package main

import (
	"fmt"

	"github.com/eskriett/confusables"
)

func main() {
	fmt.Println(confusables.ToSkeleton("𝐞х⍺𝓂𝕡Іꬲ"))
	// exarnple

	fmt.Println(confusables.IsConfusable("example", "𝐞х⍺𝓂𝕡Іꬲ"))
	// true

	fmt.Println(confusables.IsConfusable("example", "𝐞х⍺𝓂𝕡І"))
	// false
}
```
