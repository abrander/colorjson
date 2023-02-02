# (Yet Another) colorjson

colorjson is a package for formatting and colorizing JSON output
for the terminal.

Performance is not a priority of this package. It's intended
to be used for low traffic output for human consumption.

## Usage

```bash
go get github.com/abrander/colorjson
```

```go
package main

import (
	"os"

	"github.com/abrander/colorjson"
)

func main() {
	data := struct {
		Name string
		Age  int
	}{
		Name: "John Doe",
		Age:  42,
	}

	encoder := colorjson.NewEncoder(os.Stdout, colorjson.Default)
	encoder.Encode(data)
}
```
