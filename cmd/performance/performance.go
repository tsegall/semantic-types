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

type Statistic struct {
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
	Type    string
	Verbose bool
}

func main() {
	var options Options

	flag.StringVar(&options.Type, "type", "", "Select type to filter by")
	flag.BoolVar(&options.Verbose, "verbose", false, "Dump the discovery response")

	flag.Parse()

	statistics := make(map[string]Statistic)

	const FileName = 0
	const FieldOffset = 1
	const Locale = 2
	const RecordCount = 3
	const FieldName = 4
	const BaseType = 5
	const SemanticType = 6
	const Notes = 7

	ref, err := os.Open("reference.csv")
	current, err := os.Open("current.csv")

	totalRecords := 0

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
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		totalRecords++

		if recordRef[FileName] != recordCurrent[FileName] {
			log.Fatal("FileName key does not match" + recordRef[FileName])
		}
		if recordRef[FieldOffset] != recordCurrent[FieldOffset] {
			log.Fatalf("FieldOffset key does not match: %s,%s\n", recordCurrent[FileName], recordRef[FieldOffset])
		}

		// Only look for String, Boolean, Long, or Double BaseTypes where we have something to say in either the reference set or the detected set
		if (recordRef[BaseType] == "String" || recordRef[BaseType] == "Boolean" || recordRef[BaseType] == "Long" || recordRef[BaseType] == "Double") && (recordRef[SemanticType] != "" || recordCurrent[SemanticType] != "") {
			key := recordRef[FileName] + "," + recordRef[FieldOffset]
			if recordRef[BaseType] != recordCurrent[BaseType] {
				log.Printf("Key: %s - baseTypes do not match, reference: %s, current: %s\n", key, recordRef[BaseType], recordCurrent[BaseType])
			}

			if recordRef[SemanticType] == recordCurrent[SemanticType] {
				// True Positive
				update(statistics, recordRef[SemanticType], key, "", "", "")
			} else if recordRef[SemanticType] != "" && recordCurrent[SemanticType] == "" {
				// False Negative
				update(statistics, recordRef[SemanticType], "", "", "", key)
			} else if recordRef[SemanticType] == "" && recordCurrent[SemanticType] != "" {
				// False Positive
				update(statistics, recordCurrent[SemanticType], "", "", key, "")
			} else {
				update(statistics, recordRef[SemanticType], "", "", "", key)
				update(statistics, recordCurrent[SemanticType], "", "", key, "")
			}

		}

		// fmt.Println(recordRef)
		// fmt.Println(recordCurrent)
	}

	imperfectSet := make([]string, 0)
	perfectSet := make([]string, 0)

	totalTruePositives := 0
	totalFalsePositives := 0
	totalFalseNegatives := 0

	for key, element := range statistics {
		totalTruePositives += len(element.TruePositives)
		totalFalsePositives += len(element.FalsePositives)
		totalFalseNegatives += len(element.FalseNegatives)

		precision := float32(len(element.TruePositives)) / float32(len(element.TruePositives)+len(element.FalsePositives))
		recall := float32(len(element.TruePositives)) / float32(len(element.TruePositives)+len(element.FalseNegatives))
		setStatistics(statistics, key, precision, recall)

		if precision < 1.0 || recall < 1.0 {
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
		if options.Type == "" || options.Type == key {
			fmt.Printf("SemanticType: %s, Precision: %.4f, Recall: %.4f, F1 Score: %.4f (TP: %d, FP: %d, FN: %d)\n",
				key, element.Precision, element.Recall, element.F1Score, len(element.TruePositives), len(element.FalsePositives), len(element.FalseNegatives))
			if options.Verbose {
				if len(element.FalsePositives) != 0 {
					fmt.Printf("False Positives:\n")
					for _, key := range element.FalsePositives {
						fmt.Printf("\t%s\n", key)
					}
				}
				if len(element.FalseNegatives) != 0 {
					fmt.Printf("False Negatives:\n")
					for _, key := range element.FalseNegatives {
						fmt.Printf("\t%s\n", key)
					}
				}
			}
		}
	}

	fmt.Printf("\n")

	sort.Strings(perfectSet)
	for _, key := range perfectSet {
		if options.Type == "" || options.Type == key {
			fmt.Printf("SemanticType: %s, Precision: 1.0000, Recall: 1.0000, F1 Score: 1.0000 (TP: %d)\n", key, len(statistics[key].TruePositives))
		}
	}

	totalPrecision := float32(totalTruePositives) / float32(totalTruePositives+totalFalsePositives)
	totalRecall := float32(totalTruePositives) / float32(totalTruePositives+totalFalseNegatives)
	totalF1Score := 2 * ((totalPrecision * totalRecall) / (totalPrecision + totalRecall))

	if options.Type == "" {
		fmt.Printf("\nTotalPrecision: %.4f, TotalRecall: %.4f, F1 Score: %.4f (TP: %d, FP: %d, FN: %d, Record#: %d (ID%%: %.2f)\n",
			totalPrecision, totalRecall, totalF1Score, totalTruePositives, totalFalsePositives, totalFalseNegatives, totalRecords, float32((totalTruePositives+totalFalseNegatives)*100)/float32(totalRecords))
	}
}

func setStatistics(statistics map[string]Statistic, semanticType string, precision float32, recall float32) {
	val, _ := statistics[semanticType]
	val.Precision = precision
	val.Recall = recall
	val.F1Score = 2 * ((precision * recall) / (precision + recall))
	statistics[semanticType] = val
}

func update(statistics map[string]Statistic, semanticType string, tp string, tn string, fp string, fn string) {
	val, prs := statistics[semanticType]
	if !prs {
		val = Statistic{semanticType, make([]string, 0), make([]string, 0), make([]string, 0), make([]string, 0), 0.0, 0.0, 0.0}
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
