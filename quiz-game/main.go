package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"time"
)

var (
	// default values for flags
	duration   = 30
	correctCnt = 0
	fileName   = "problems.csv"
)

func main() {
	flag.IntVar(&duration, "d", duration, "configuration time limit the quiz")
	flag.StringVar(&fileName, "f", fileName, "file with problems/answers for the quiz")
	flag.Parse()

	// open
	f, err := os.Open(fileName)
	if err != nil {
		panic(err)
	}

	// read
	c := csv.NewReader(f)
	csv, err := c.ReadAll()
	if err != nil {
		panic(err)
	}

	timer := time.NewTimer(time.Duration(duration) * time.Second)
	ansChan := make(chan string)
	for i, v := range csv {
		fmt.Printf("Problem #%v: %v = ", i+1, v[0])
		go func(out chan<- string) {
			var answer string
			fmt.Scanln(&answer)
			out <- answer
		}(ansChan)

		select {
		case <-timer.C:
			fmt.Printf("\nYou run out of time! You have answered correctly for %v out of %v\n", correctCnt, len(csv))
			return
		case ans := <-ansChan:
			if ans == v[1] {
				correctCnt++
			}
		}
	}

	// my solution of the quiz game
	// // timer
	// timer := time.AfterFunc(time.Duration(duration)*time.Second, func() {
	// 	fmt.Printf("\nYou run out of time! You have answered correctly for %v out of %v\n", correctCnt, len(csv))
	// 	os.Exit(0)
	// })
	// for i, v := range csv {
	// 	var answer string
	// 	fmt.Printf("Problem #%v: %v = ", i+1, v[0])
	// 	fmt.Scanln(&answer)
	// 	if answer == v[1] {
	// 		correctCnt++
	// 	}
	// }
	// fmt.Printf("You have answered correctly for %v out of %v\n", correctCnt, len(csv))
	// timer.Stop()
}
