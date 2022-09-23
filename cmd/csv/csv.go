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
	FieldOffset string
	FileName    string
}

func main() {
	var options Options

	const FileName = 0
	const FieldOffset = 1
	const Locale = 2
	const RecordCount = 3
	const FieldName = 4
	const BaseType = 5
	const SemanticType = 6
	const Notes = 7

	flag.StringVar(&options.Columns, "column", "", "Columns to extract")
	flag.StringVar(&options.Columns, "c", "", "Columns to extract")
	flag.StringVar(&options.FieldOffset, "fieldoffset", "-1", "Field offset")
	flag.StringVar(&options.FileName, "filename", "", "File name")
	flag.Parse()

	_, err := strconv.Atoi(options.FieldOffset)
	if err != nil {
		log.Fatal(err)
	}

	ref, err := os.Open("reference.csv")
	if err != nil {
		log.Fatal(err)
	}

	columnsString := strings.Split(options.Columns, ",")
	columnsInt := make([]int, len(columnsString))

	for i := 0; i < len(columnsInt); i++ {
		columnsInt[i], err = strconv.Atoi(columnsString[i])
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

		if options.FileName != "" && recordRef[FileName] != options.FileName {
			continue
		}
		if options.FieldOffset != "-1" && recordRef[FieldOffset] != options.FieldOffset {
			continue
		}

		if len(columnsInt) != 0 {
			first := true
			for i := 0; i < len(columnsInt); i++ {
				if first {
					first = false
				} else {
					fmt.Print(",")
				}
				fmt.Print(recordRef[columnsInt[i]])
			}
			fmt.Println()
		} else {
			fmt.Println(recordRef)
		}
	}
}
