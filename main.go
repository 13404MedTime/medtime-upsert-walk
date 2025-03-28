package main

import (
	"fmt"
	"time"
)

func main() {
	t := "2023-12-28T13:33:00.000Z"
	parse, _ := time.Parse(time.RFC3339, t)
	fmt.Println(parse)
}
