package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func readFromFile(filename string) {
	f,err :=os.Open(filename)
	if err != nil {
		fmt.Println(err)
	}
	defer f.Close()

	scanner :=bufio.NewScanner(f)
	words := []string{}

	for scanner.Scan(){
		words = strings.Split(scanner.Text(), " ")
		for i,word :=range words {
			fmt.Println(i,word)
		}
	}
}
