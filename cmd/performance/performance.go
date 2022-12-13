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
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
)

type SimpleStatistic struct {
	BaseType string
	Count    int
}

type SemanticStatistic struct {
	SemanticType   string
	TruePositives  []string
	FalsePositives []string
	TrueNegatives  []string
	FalseNegatives []string
	Precision      float32
	Recall         float32
	F1Score        float32
}

type Options struct {
	BaseType     bool
	Locale       string
	SemanticType string
	Verbose      bool
}

const FileNameIndex = 0
const FieldOffsetIndex = 1
const LocaleIndex = 2
const RecordCountIndex = 3
const FieldNameIndex = 4
const BaseTypeIndex = 5
const TypeModifierIndex = 6
const SemanticTypeIndex = 7
const NotesIndex = 8

func main() {
	var options Options

	flag.BoolVar(&options.BaseType, "baseType", false, "Output Base Type information")
	flag.StringVar(&options.Locale, "locale", "", "Select Locale to filter by")
	flag.StringVar(&options.SemanticType, "semanticType", "", "Select Semantic Type to filter by")
	flag.BoolVar(&options.Verbose, "verbose", false, "Dump the discovery response")

	flag.Parse()

	simpleStatistics := make(map[string]*SimpleStatistic)
	statistics := make(map[string]*SemanticStatistic)

	ref, err := os.Open("reference.csv")
	current, err := os.Open("current.csv")

	totalRecords := 0
	totalDataRecords := 0
	totalBaseTypeErrors := 0

	if err != nil {
		log.Fatal(err)
	}

	defer ref.Close()
	defer current.Close()

	readerRef := bufio.NewReader(ref)
	readerCurrent := bufio.NewReader(current)

	csvRef := csv.NewReader(readerRef)
	csvRef.Read()
	csvCurrent := csv.NewReader(readerCurrent)
	csvCurrent.Read()

	for {
		recordRef, err := csvRef.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		recordCurrent, err := csvCurrent.Read()
		if err == io.EOF {
			log.Fatalf("Short read on Current file, record #: %d\n", totalRecords)
		}
		if err != nil {
			log.Fatal(err)
		}

		// Check to see if we are processing a single Locale
		if options.Locale != "" && options.Locale != recordRef[LocaleIndex] {
			continue;
		}

		totalRecords++
		if recordRef[TypeModifierIndex] != "NULL" && recordRef[TypeModifierIndex] != "BLANK" && recordRef[TypeModifierIndex] != "BLANKORNULL" {
			totalDataRecords++
		}

		baseType := recordCurrent[BaseTypeIndex]
		val, exists := simpleStatistics[baseType]
		if exists {
			val.Count++
		} else {
			simpleStatistics[baseType] = &SimpleStatistic{baseType, 1}
		}

		if recordRef[FileNameIndex] != recordCurrent[FileNameIndex] {
			log.Fatal("FileName key does not match" + recordRef[FileNameIndex])
		}
		if recordRef[FieldOffsetIndex] != recordCurrent[FieldOffsetIndex] {
			log.Fatalf("FieldOffset key does not match: %s,%s\n", recordCurrent[FileNameIndex], recordRef[FieldOffsetIndex])
		}

		key := recordRef[FileNameIndex] + "," + recordRef[FieldOffsetIndex]
		if options.BaseType && recordRef[BaseTypeIndex] != recordCurrent[BaseTypeIndex] {
			log.Printf("Key: %s (%s) - baseTypes do not match, reference: %s, current: %s\n", key, recordRef[FieldNameIndex], recordRef[BaseTypeIndex], recordCurrent[BaseTypeIndex])
			totalBaseTypeErrors++
		}

		// Only look for String, Boolean, Long, or Double BaseTypes where we have something to say in either the reference set or the detected set
		if (recordRef[BaseTypeIndex] == "String" || recordRef[BaseTypeIndex] == "Boolean" || recordRef[BaseTypeIndex] == "Long" || recordRef[BaseTypeIndex] == "Double") && (recordRef[SemanticTypeIndex] != "" || recordCurrent[SemanticTypeIndex] != "") {
			if recordRef[BaseTypeIndex] != recordCurrent[BaseTypeIndex] {
				log.Printf("Key: %s - baseTypes do not match, reference: %s, current: %s\n", key, recordRef[BaseTypeIndex], recordCurrent[BaseTypeIndex])
			}

			if recordRef[SemanticTypeIndex] == recordCurrent[SemanticTypeIndex] {
				// True Positive
				update(statistics, recordRef[SemanticTypeIndex], key, "", "", "")
			} else if recordRef[SemanticTypeIndex] != "" && recordCurrent[SemanticTypeIndex] == "" {
				// False Negative
				update(statistics, recordRef[SemanticTypeIndex], "", "", "", key)
			} else if recordRef[SemanticTypeIndex] == "" && recordCurrent[SemanticTypeIndex] != "" {
				// False Positive
				update(statistics, recordCurrent[SemanticTypeIndex], "", "", key, "")
			} else {
				update(statistics, recordRef[SemanticTypeIndex], "", "", "", key)
				update(statistics, recordCurrent[SemanticTypeIndex], "", "", key, "")
			}

		}

		// fmt.Println(recordRef)
		// fmt.Println(recordCurrent)
	}

	imperfectSet := make([]string, 0)
	perfectSet := make([]string, 0)
	notDetectedSet := make([]string, 0)

	totalTruePositives := 0
	totalFalsePositives := 0
	totalFalseNegatives := 0
	notDetected := 0

	for key, element := range statistics {
		if len(element.TruePositives) == 0 && len(element.FalsePositives) == 0 {
			notDetected += len(element.FalseNegatives)
		} else {
			totalTruePositives += len(element.TruePositives)
			totalFalsePositives += len(element.FalsePositives)
			totalFalseNegatives += len(element.FalseNegatives)
		}

		precision := float32(len(element.TruePositives)) / float32(len(element.TruePositives)+len(element.FalsePositives))
		recall := float32(len(element.TruePositives)) / float32(len(element.TruePositives)+len(element.FalseNegatives))
		setStatistics(statistics, key, precision, recall)

		if len(element.TruePositives) == 0 && len(element.FalsePositives) == 0 {
			notDetectedSet = append(notDetectedSet, key)
		} else if precision < 1.0 || recall < 1.0 {
			imperfectSet = append(imperfectSet, key)
		} else {
			perfectSet = append(perfectSet, key)
		}
	}

	sort.SliceStable(imperfectSet, func(i, j int) bool {
		return statistics[imperfectSet[i]].F1Score < statistics[imperfectSet[j]].F1Score
	})

	for _, key := range imperfectSet {
		element := statistics[key]
		if options.SemanticType == "" || options.SemanticType == key {
			fmt.Printf("SemanticType: %s, Precision: %.4f, Recall: %.4f, F1 Score: %.4f (TP: %d, FP: %d, FN: %d)\n",
				key, element.Precision, element.Recall, element.F1Score, len(element.TruePositives), len(element.FalsePositives), len(element.FalseNegatives))
			if options.Verbose {
				if len(element.FalsePositives) != 0 {
					for _, key := range element.FalsePositives {
						fmt.Printf("FP\t%s\n", key)
					}
				}
				if len(element.FalseNegatives) != 0 {
					for _, key := range element.FalseNegatives {
						fmt.Printf("FN\t%s\n", key)
					}
				}
			}
		}
	}

	fmt.Printf("\n")

	sort.Strings(perfectSet)
	for _, key := range perfectSet {
		if options.SemanticType == "" || options.SemanticType == key {
			fmt.Printf("SemanticType: %s, Precision: 1.0000, Recall: 1.0000, F1 Score: 1.0000 (TP: %d)\n", key, len(statistics[key].TruePositives))
		}
	}

	fmt.Printf("\n")

	sort.Strings(notDetectedSet)
	for _, key := range notDetectedSet {
		if options.SemanticType == "" || options.SemanticType == key {
			fmt.Printf("SemanticType: %s, Precision: 0.0000, Recall: NaN, F1 Score: NaN (FN: %d)\n", key, len(statistics[key].FalseNegatives))
		}
	}

	totalPrecision := float32(totalTruePositives) / float32(totalTruePositives+totalFalsePositives)
	totalRecall := float32(totalTruePositives) / float32(totalTruePositives+totalFalseNegatives)
	totalF1Score := 2 * ((totalPrecision * totalRecall) / (totalPrecision + totalRecall))

	if options.SemanticType == "" {
		fmt.Printf("\nSemantic Types: %d, TotalPrecision: %.4f, TotalRecall: %.4f, F1 Score: %.4f (TP: %d, FP: %d, FN: %d), NotDetected: %d, Record# (Non-null/blank): %d (%d) (ID%%: %.2f)\n",
			len(statistics) - len(notDetectedSet), totalPrecision, totalRecall, totalF1Score, totalTruePositives, totalFalsePositives, totalFalseNegatives, notDetected,
			totalRecords, totalDataRecords, float32((totalTruePositives+totalFalseNegatives)*100)/float32(totalDataRecords))
	}

	if options.BaseType {
		fmt.Printf("Base Types: ")
		index := 0
		totalBaseTypes := 0
		for _, baseType := range simpleStatistics {
			if index != 0 {
				fmt.Print(", ")
			}
			index++
			totalBaseTypes += baseType.Count
			fmt.Printf("%s: %d", baseType.BaseType, baseType.Count)
		}
		fmt.Printf(" (Errors: %.4f%% (%d))\n", float32(totalBaseTypeErrors*100)/float32(totalBaseTypes), totalBaseTypeErrors)
	}
}

func setStatistics(statistics map[string]*SemanticStatistic, semanticType string, precision float32, recall float32) {
	statistics[semanticType].Precision = precision
	statistics[semanticType].Recall = recall
	statistics[semanticType].F1Score = 2 * ((precision * recall) / (precision + recall))
}

func update(statistics map[string]*SemanticStatistic, semanticType string, tp string, tn string, fp string, fn string) {
	val, prs := statistics[semanticType]
	if !prs {
		val = &SemanticStatistic{semanticType, make([]string, 0), make([]string, 0), make([]string, 0), make([]string, 0), 0.0, 0.0, 0.0}
	}

	if tp != "" {
		val.TruePositives = append(val.TruePositives, tp)
	}
	if fp != "" {
		val.FalsePositives = append(val.FalsePositives, fp)
	}
	if tn != "" {
		val.TrueNegatives = append(val.TrueNegatives, tn)
	}
	if fn != "" {
		val.FalseNegatives = append(val.FalseNegatives, fn)
	}
	statistics[semanticType] = val
}
