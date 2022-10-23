package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)

type Options struct {
	Columns     string
	FileName    string
}

func main() {
	var options Options

	flag.StringVar(&options.Columns, "column", "", "Columns to extract")
	flag.StringVar(&options.Columns, "c", "", "Columns to extract")
	flag.StringVar(&options.FileName, "filename", "reference.csv", "File name")
	flag.Parse()

	ref, err := os.Open(options.FileName)
	if err != nil {
		log.Fatal(err)
	}

	columnsString := strings.Split(options.Columns, ",")
	columnsInt := make([]int, len(columnsString))
	quotes := make([]bool, len(columnsString))

	for i := 0; i < len(columnsInt); i++ {
		index := strings.Index(columnsString[i], ":")
		if index == -1 {
			columnsInt[i], err = strconv.Atoi(columnsString[i])
		} else {
			columnsInt[i], err = strconv.Atoi(columnsString[i][0:len(columnsString[i])-2])
			quotes[i] = true
		}
		if err != nil {
			log.Fatal(err)
		}
	}

	defer ref.Close()

	readerRef := bufio.NewReader(ref)

	csvRef := csv.NewReader(readerRef)
	csvRef.Read()

	for {
		recordRef, err := csvRef.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		if len(columnsInt) != 0 {
			first := true
			for i := 0; i < len(columnsInt); i++ {
				if first {
					first = false
				} else {
					fmt.Print(",")
				}
				if quotes[i] {
					fmt.Print("\"")
				}
				fmt.Print(recordRef[columnsInt[i]])
				if quotes[i] {
					fmt.Print("\"")
				}
			}
			fmt.Println()
		} else {
			fmt.Println(recordRef)
		}
	}
}
