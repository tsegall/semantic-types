/*
 * Copyright 2022 Tim Segall
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
)

type Statistic struct {
	SemanticType     string `json:"semanticType"`
	CorrelationCount []int `json:"correlationCount"`
	Correlation      []float32 `json:"correlation"`
	Direction        []int `json:"direction"`
	TotalDistance    []int `json:"totalDistance"`
	TotalMatches     int `json:"totalMatches"`
	Index            int `json:"index"`
}

func getByIndex(statistics map[string]*Statistic, searching int) *Statistic {
	for _, value := range statistics {
		if value.Index == searching {
			return value
		}
	}

	return nil
}

func isSignificant(statistics map[string]*Statistic, key string) bool {
	s := statistics[key]
	if s.TotalMatches < 10 {
		return false
	}

	interesting := false
	for index, value := range s.Correlation {
		if value > .1 && getByIndex(statistics, index).TotalMatches >= 10 {
			interesting = true
			// log.Printf("*** Semantic Type: %s, index: %d (%s), value: %.2f\n", s.SemanticType, index, getByIndex(statistics, index).SemanticType, value)
			break
		}
	}
	return interesting && s.TotalMatches > 10
}

type Options struct {
	Locale      string
	Correlation bool
	Direction   bool
	Distance    bool
	Format      string
	Totals      bool
	Verbose     bool
}

const FileName = 0
const FieldOffset = 1
const Locale = 2
const RecordCount = 3
const FieldName = 4
const BaseType = 5
const TypeModifier = 6
const SemanticType = 7
const Notes = 8

func main() {
	var options Options

	flag.BoolVar(&options.Correlation, "correlation", true, "Output the Correlation matrix")
	flag.BoolVar(&options.Direction, "direction", false, "Output the Direction matrix")
	flag.BoolVar(&options.Distance, "distance", false, "Output the Distance matrix")
	flag.StringVar(&options.Format, "format", "human", "Set the output format")
	flag.StringVar(&options.Locale, "locale", "en-US", "Set the locale")
	flag.BoolVar(&options.Totals, "totals", false, "Output the Totals facts")
	flag.BoolVar(&options.Verbose, "verbose", false, "Additional debugging information")

	flag.Parse()

	statistics := make(map[string]*Statistic)

	ref, err := os.Open("reference.csv")

	if err != nil {
		log.Fatal(err)
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

		if recordRef[Locale] != options.Locale {
			continue
		}

		// Only look for String, Boolean, Long, or Double BaseTypes where we have something to say in either the reference set or the detected set
		if (recordRef[BaseType] == "String" || recordRef[BaseType] == "Boolean" || recordRef[BaseType] == "Long" || recordRef[BaseType] == "Double") && recordRef[SemanticType] != "" {
			semanticType := recordRef[SemanticType]
			_, exists := statistics[semanticType]
			if !exists {
				statistics[recordRef[SemanticType]] = &Statistic{SemanticType: semanticType}
			}
		}
	}

	// Grab all the Semantic Types and sort them
	keys := make([]string, 0, len(statistics))
	for key := range statistics {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool { return statistics[keys[i]].SemanticType < statistics[keys[j]].SemanticType })

	calculateFacts(statistics, options, keys)

	if options.Correlation {
		correlationOutput(statistics, options, keys)
	}

	if options.Distance {
		// Output the Distance header
		for _, key := range keys {
			fmt.Printf(",%s", key)
		}
		fmt.Println()
		// Output the Distance matrix
		for i := 0; i < len(statistics); i++ {
			fmt.Print(keys[i])
			for _, key := range keys {
				fmt.Print(",")
				if statistics[key].TotalMatches > 0 {
					fmt.Printf("%.2f", float32(statistics[key].TotalDistance[i])/float32(statistics[key].TotalMatches))
				}
			}
			fmt.Println()
		}
	}

	if options.Direction {
		// Output the Left/Right header
		for _, key := range keys {
			fmt.Printf(",%s", key)
		}
		fmt.Println()
		// Output the Left/Right matrix
		for i := 0; i < len(statistics); i++ {
			fmt.Print(keys[i])
			for _, key := range keys {
				fmt.Print(",")
				if statistics[key].TotalMatches > 0 {
					fmt.Printf("%.2f", float32(statistics[key].Direction[i])/float32(statistics[key].TotalMatches))
				}
			}
			fmt.Println()
		}
	}

	if options.Totals {
		for _, key := range keys {
			fmt.Printf("%s: %d\n", key, statistics[key].TotalMatches)
		}
		fmt.Println()
	}
}

func calculateFacts(statistics map[string]*Statistic, options Options, keys []string) {
	ref, err := os.Open("reference.csv")
	if err != nil {
		log.Fatal(err)
	}

	// Now we know how many Semantic Types we have allocate space and set the Index
	for index, key := range keys {
		statistics[key].CorrelationCount = make([]int, len(statistics))
		statistics[key].Correlation = make([]float32, len(statistics))
		statistics[key].Direction = make([]int, len(statistics))
		statistics[key].TotalDistance = make([]int, len(statistics))
		statistics[key].Index = index
	}

	readerRef := bufio.NewReader(ref)

	csvRef := csv.NewReader(readerRef)
	csvRef.Read()

	currentFileName := ""
	currentMap := make([]string, 0)

	for {
		recordRef, err := csvRef.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		if recordRef[Locale] != options.Locale {
			continue
		}

		if currentFileName == "" {
			currentFileName = recordRef[FileName]
		}
		currentMap = append(currentMap, recordRef[SemanticType])
		if recordRef[FileName] != currentFileName {
			// Iterate through all the Semantic Types for this file
			for index, value := range currentMap {
				// If this field has a Semantic Type
				seen := mapset.NewSet[string]()
				if value != "" {
					statistics[value].TotalMatches++
					// Look at the whole record for other Semantic types to generate the correlation
					for indexCorr, valueCorr := range currentMap {
						// If this field has a Semantic Type and it is not the current field and it is not the same as the one we are searching for
						if valueCorr != "" && index != indexCorr && value != valueCorr {
							if seen.Contains(valueCorr) {
								continue
							}
							seen.Add(valueCorr)
							semanticOffset := statistics[valueCorr].Index
							statistics[value].TotalDistance[semanticOffset] += Abs(indexCorr - index)
							if indexCorr < index {
								statistics[value].Direction[semanticOffset]++
							} else {
								statistics[value].Direction[semanticOffset]--
							}
							// fmt.Printf("%s -> %s: %d -> %d(%d) (%f)\n", value, valueCorr, index, indexCorr, semanticOffset, correlation)
							statistics[value].CorrelationCount[semanticOffset]++
						}
					}
				}
			}
			// log.Printf("Dumping '%s', '%d'\n", currentFileName, currentFieldOffset)
			currentMap = make([]string, 0)
			currentFileName = recordRef[FileName]
		}

	}
	// TODO grab last one

	for i := 0; i < len(statistics); i++ {
		for _, key := range keys {
			statistics[key].Correlation[i] = float32(statistics[key].CorrelationCount[i]) / float32(statistics[key].TotalMatches)
		}
	}
}

func correlationOutput(statistics map[string]*Statistic, options Options, keys []string) {
	if options.Format == "human" {
		// Output the Correlation header
		for _, key := range keys {
			if isSignificant(statistics, key) {
				fmt.Printf(",%s", key)
			}
		}
		fmt.Println()
		// Output the Correlation matrix
		for i := 0; i < len(statistics); i++ {
			if isSignificant(statistics, keys[i]) {
				fmt.Print(keys[i])
				for _, key := range keys {
					if isSignificant(statistics, key) {
						fmt.Printf(",%.2f", float32(statistics[key].CorrelationCount[i])/float32(statistics[key].TotalMatches))
					}
				}
				fmt.Println()
			}
		}
	} else if options.Format == "data" {
		locale := strings.Replace(options.Locale, "-", "_", -1)
		fileName := "SemanticTypes__" + locale + ".json"

		bytes, err := json.Marshal(Values(statistics))
		check(err)

		ioutil.WriteFile(fileName, bytes, 0644)
	}
}

func Values[M ~map[K]V, K comparable, V any](m M) []V {
    r := make([]V, 0, len(m))
    for _, v := range m {
        r = append(r, v)
    }
    return r
}

func Abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
