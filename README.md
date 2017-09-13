## ENVI
Environment vars in Go

## Example

A very basic example (check the `examples` folder):

```go
package main

import (
	"fmt"
	"github.com/douglasmakey/envi"
)

type environments struct {
	Intent int            `env:"INTENT"`
	Ports  []int          `env:"PORTS" envDefault:"3000"`
	IsProd bool           `env:"PROD,required"`
	IsDev  bool           `env:"DEV"`
	Hosts  []string       `env:"HOSTS" envSeparator:":"`
	Sector map[string]int `env:"SECTOR"`
}

func main() {
	env := environments{}
	err := envi.Parse(&env)
	if err != nil {
		panic(fmt.Sprintf("%+v\n", err))
	}

	fmt.Printf("%+v\n", env)
}
```

You can run it like this:

```sh
$ NTENT=5 PROD=true HOSTS="127.0.0.1:localhost" SECTOR="a:1,b:2,c:4"  go run examples/examples.go
Intent:5 Ports:[3 0 0 0] IsProd:true IsDev:false Hosts:[127.0.0.1 localhost] Sector:map[a:1 b:2 c:4]}
```

## Supported types and defaults

This library has support for the following types:

* `string`
* `int`
* `uint`
* `int64`
* `bool`
* `float32`
* `float64`
* `map`
* `[]string`
* `[]int`
* `[]bool`
* `[]float32`
* `[]float64`


You can set the `envDefault` tag for something, this value will be used in the
case of absence of it in the environment. If you don't do that AND the
environment variable is also not set, the zero-value
of the type will be used: empty for `string`s, `false` for `bool`s
and `0` for `int`s.

By default, slice types will split the environment value on `,`; you can change this behavior by setting the `envSeparator` tag.

## Required fields

The `env` tag option `required` (e.g., `env:"MyKey,required"`) can be added
to ensure that some environment variable is set.
