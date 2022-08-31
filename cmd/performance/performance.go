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
	TruePositives  int
	FalsePositives int
	TrueNegatives  int
	FalseNegatives int
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

		if recordRef[FileName] != recordCurrent[FileName] {
			log.Fatal("FileName key does not match" + recordRef[FileName])
		}
		if recordRef[FieldOffset] != recordCurrent[FieldOffset] {
			log.Fatal("FieldOffset key does not match" + recordRef[FieldOffset])
		}

		// Only look for String, Long, or Double BaseTypes where we have something to say in either the reference set or the detected set
		if (recordRef[BaseType] == "String" || recordRef[BaseType] == "Long" || recordRef[BaseType] == "Double") && (recordRef[SemanticType] != "" || recordCurrent[SemanticType] != "") {
			if recordRef[SemanticType] == recordCurrent[SemanticType] {
				// True Positive
				update(statistics, recordRef[SemanticType], 1, 0, 0, 0)
			} else if recordRef[SemanticType] != "" && recordCurrent[SemanticType] == "" {
				// False Negative
				update(statistics, recordRef[SemanticType], 0, 0, 0, 1)
			} else if recordRef[SemanticType] == "" && recordCurrent[SemanticType] != "" {
				// False Positive
				update(statistics, recordCurrent[SemanticType], 0, 0, 1, 0)
			} else {
				update(statistics, recordRef[SemanticType], 0, 0, 0, 1)
				update(statistics, recordCurrent[SemanticType], 0, 0, 1, 0)
			}

		}

		// fmt.Println(recordRef)
		// fmt.Println(recordCurrent)
	}

	imperfectSet := make([]string, 0)
	perfectSet := make([]string, 0)

	for key, element := range statistics {
		precision := float32(element.TruePositives) / float32(element.TruePositives+element.FalsePositives)
		recall := float32(element.TruePositives) / float32(element.TruePositives+element.FalseNegatives)
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
				key, element.Precision, element.Recall, element.F1Score, element.TruePositives, element.FalsePositives, element.FalseNegatives)
		}
	}

	fmt.Printf("\n")

	sort.Strings(perfectSet)
	for _, key := range perfectSet {
		if options.Type == "" || options.Type == key {
			fmt.Printf("SemanticType: %s, Precision: 1.0, Recall: 1.0, F1 Score: 1.0 (TP: %d)\n", key, statistics[key].TruePositives)
		}
	}
}

func setStatistics(statistics map[string]Statistic, semanticType string, precision float32, recall float32) {
	val, _ := statistics[semanticType]
	val.Precision = precision
	val.Recall = recall
	val.F1Score = 2 * ((precision * recall) / (precision + recall))
	statistics[semanticType] = val
}

func update(statistics map[string]Statistic, semanticType string, tp int, tn int, fp int, fn int) {
	val, prs := statistics[semanticType]
	if prs {
		val.TruePositives += tp
		val.FalsePositives += fp
		val.TrueNegatives += tn
		val.FalseNegatives += fn
		statistics[semanticType] = val
	} else {
		statistics[semanticType] = Statistic{semanticType, tp, fp, tn, fn, 0.0, 0.0, 0.0}
	}
}
