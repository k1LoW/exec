package main

import (
	"flag"
	"time"
)

func main() {
	var (
		sleep = flag.Int("sleep", 0, "sleep sec")
	)
	flag.Parse()

	if *sleep > 0 {
		time.Sleep(time.Duration(*sleep) * time.Second)
	}
}
