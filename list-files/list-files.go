package main

import (
	"fmt"
	"log"
	"os"
)

func listfiles(dir string, ch chan string) {
	defer close(ch)
	files, err := os.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
		return
	}

	var chlist []chan string
	for _, file := range files {
		ch <- dir + "/" + file.Name()
		if file.IsDir() {
			c := make(chan string)
			chlist = append(chlist, c)
			go listfiles(dir+"/"+file.Name(), c)
		}
	}

	for i := range chlist {
		for c := range chlist[i] {
			ch <- c
		}
	}
}

func main() {
	ch := make(chan string)
	go listfiles("readable", ch)
	for s := range ch {
		fmt.Println(s)
	}
}
