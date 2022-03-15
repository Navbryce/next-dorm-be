package util

import (
	"fmt"
	"math/rand"
)

var names = []string{
	"Dog",
	"Cat",
	"Frog",
	"Wreck",
}

func GenerateAlias() string {
	return fmt.Sprintf("Anon %v", names[rand.Intn(len(names))])
}
