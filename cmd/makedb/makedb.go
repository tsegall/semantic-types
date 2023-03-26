package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
)

type FieldAnalysis struct {
	FieldName          string   `json:"fieldName"`
	TotalCount         int      `json:"totalCount"`
	SampleCount        int      `json:"sampleCount"`
	MatchCount         int      `json:"matchCount"`
	NullCount          int      `json:"nullCount"`
	BlankCount         int      `json:"blankCount"`
	DistinctCount      int      `json:"distinctCount"`
	RegExp             string   `json:"regExp"`
	Confidence         float64  `json:"confidence"`
	Type               string   `json:"type"`
	TypeModifier       string   `json:"typeModifier"`
	Min                string   `json:"min"`
	Max                string   `json:"max"`
	MinLength          int      `json:"minLength"`
	MaxLength          int      `json:"maxLength"`
	TopK               []string `json:"topK"`
	BottomK            []string `json:"bottomK"`
	Cardinality        int      `json:"cardinality"`
	OutlierCardinality int      `json:"outlierCardinality"`
	ShapesCardinality  int      `json:"shapesCardinality"`
	Percentiles        []string `json:"percentiles"`
	LeadingWhiteSpace  bool     `json:"leadingWhiteSpace"`
	TrailingWhiteSpace bool     `json:"trailingWhiteSpace"`
	Multiline          bool     `json:"multiline"`
	DateResolutionMode string   `json:"dateResolutionMode"`
	IsSemanticType     bool     `json:"isSemanticType"`
	SemanticType       string   `json:"semanticType"`
	LogicalType        bool     `json:"logicalType"`
	TypeQualifier      string   `json:"typeQualifier"`
	KeyConfidence      float64  `json:"keyConfidence"`
	Uniqueness         float64  `json:"uniqueness"`
	DetectionLocale    string   `json:"detectionLocale"`
	FtaVersion         string   `json:"ftaVersion"`
	StructureSignature string   `json:"structureSignature"`
	DataSignature      string   `json:"dataSignature"`
}

func main() {
	flag.Parse()

	for _, fileName := range flag.Args() {
		file, _ := ioutil.ReadFile(fileName)

		var fields []FieldAnalysis

		err := json.Unmarshal([]byte(file), &fields)
		if err != nil {
			fmt.Println("error:", err)
		}

		// File,FieldOffset,Locale,RecordCount,FieldName,BaseType,TypeModifier,SemanticType,Notes
		outputName := fileName
		if strings.HasSuffix(outputName, ".out") {
			outputName = outputName[:len(fileName)-4]
		}
		for i := 0; i < len(fields); i++ {
			var major int
			if fields[i].FtaVersion == ""  {
				fmt.Printf("****** WARNING NO VERSION ******");
				major = 12;
			} else {
				versions := strings.Split(fields[i].FtaVersion, ".")
				major, _ = strconv.Atoi(versions[0])
			}
			var semanticType string
			var typeModifier string
			if (major >= 12) {
				semanticType = fields[i].SemanticType
				typeModifier = fields[i].TypeModifier
			} else {
				if fields[i].LogicalType {
					semanticType = fields[i].TypeQualifier
				} else {
					typeModifier = fields[i].TypeQualifier
				}
			}
			fmt.Printf(`%s,%d,"%s",%d,"%s","%s","%s","%s","%s"`,
				outputName, i, fields[i].DetectionLocale, fields[i].SampleCount, fields[i].FieldName, fields[i].Type, typeModifier, semanticType, "")
			fmt.Println()
		}
	}
}
