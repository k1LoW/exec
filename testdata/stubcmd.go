package main

import (
	"flag"
	"fmt"
	"time"
)

func main() {
	var (
		sleep = flag.Int("sleep", 0, "sleep sec")
		echo  = flag.String("echo", "", "echo string")
	)
	flag.Parse()

	if *sleep > 0 {
		time.Sleep(time.Duration(*sleep) * time.Second)
	}

	if *echo != "" {
		fmt.Printf("%s\n", *echo)
	}
}
