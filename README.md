# confusables

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
