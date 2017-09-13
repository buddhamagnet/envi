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
	err = envi.ChangeValue("Intent", "33")
	if err != nil {
		panic(fmt.Sprintf("%+v\n", err))
	}

	fmt.Printf("%+v\n", env)
}
