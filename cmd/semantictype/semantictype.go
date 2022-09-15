package main

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"cobber.com/internal/coretypes"
)

type Options struct {
	Download     bool
	Column       string
	Source       string
	MaxColumns   int
	MaxDownloads int
	MinLines     int
	Verbose      bool
}

func main() {
	var options Options

	flag.BoolVar(&options.Download, "download", false, "Download files after discovery")
	flag.StringVar(&options.Column, "column", "", "Column name to search for")
	flag.StringVar(&options.Source, "source", "socrata", "Data source (socrata, data.sfgov.org")
	flag.IntVar(&options.MaxColumns, "maxColumns", 40, "Maximum number of columns (-1 for unlimited)")
	flag.IntVar(&options.MaxDownloads, "maxDownloads", 10, "Maximum number of files to download (-1 for unlimited)")
	flag.IntVar(&options.MinLines, "minLines", 20, "Minimum number of data lines in a file to be interesting")
	flag.BoolVar(&options.Verbose, "verbose", false, "Dump the discovery response")

	flag.Parse()

	var dataDirectory string

	var discovery string

	if options.Source == "socrata" {
		discovery = "http://api.us.socrata.com/api/catalog/v1"
		dataDirectory = "data/opendata_socrata_com"
	} else if options.Source == "data.sfgov.org" {
		discovery = "http://data.sfgov.org/api/catalog/v1"
		dataDirectory = "data/data_sfgov_org"
	} else {
		panic("Unknown source: " + options.Source)
	}

	c := http.Client{Timeout: time.Duration(10) * time.Second}
	req, err := http.NewRequest("GET", discovery, nil)
	if err != nil {
		fmt.Printf("error %s", err)
		return
	}
	req.Header.Add("Accept", `application/json`)

	// if you appending to existing query this works fine
	q := req.URL.Query()
	if options.Column != "" {
		q.Add("column_names", options.Column)
	}
	q.Add("public", "true")
	q.Add("limit", "100")

	var datasets []coretypes.DataSet
	lastResourceID := ""
	retrieved := 0
	for {
		q.Set("scroll_id", lastResourceID)
		req.URL.RawQuery = q.Encode()

		fmt.Println(req.URL.String())

		resp, err := c.Do(req)
		if err != nil {
			fmt.Printf("error %s", err)
			return
		}

		body, _ := ioutil.ReadAll(resp.Body)

		var result coretypes.DiscoveryResponse
		if err := json.Unmarshal(body, &result); err != nil { // Parse []byte to go struct pointer
			fmt.Println("Cannot unmarshal JSON")
		}
		resp.Body.Close()

		if options.Verbose {
			fmt.Printf("%s", body)
		}

		for _, rec := range result.Results {
			if rec.Resource.DownloadCount > 20 {
				datasets = append(datasets, coretypes.DataSet{rec.Metadata.Domain, rec.Resource.ID})
			}
			retrieved++
			lastResourceID = rec.Resource.ID
		}

		if retrieved >= result.ResultSetSize {
			break
		}
	}

	fmt.Printf("Located %d items\n", len(datasets))

	var saveCount = 0
	for _, dataset := range datasets {
		directory := dataDirectory + "/" + dataset.Host
		if _, err := os.Stat(directory); os.IsNotExist(err) {
			os.MkdirAll(directory, 0750)
		}

		if options.Download && saveCount < options.MaxDownloads {
			if save(c, dataset, directory, options) {
				saveCount++
			}
		}
	}
}

func save(c http.Client, dataset coretypes.DataSet, directory string, options Options) bool {
	filename := directory + "/" + dataset.ID + ".csv"
	// Skip if the file already exists
	if _, err := os.Stat(filename); !errors.Is(err, os.ErrNotExist) {
		fmt.Printf("Already downloaded: %s\n", filename)
		return false
	}

	req, err := http.NewRequest("GET", "https://"+dataset.Host+"/resource/"+dataset.ID+".csv", nil)
	if err != nil {
		fmt.Printf("error %s", err)
		return false
	}

	resp, err := c.Do(req)
	if err != nil {
		fmt.Printf("error %s", err)
		return false
	}

	defer resp.Body.Close()
	file, err := os.Create(filename)
	if err != nil {
		panic(err)
	}

	file.ReadFrom(resp.Body)
	file.Close()

	file, _ = os.Open(filename)
	fileScanner := bufio.NewScanner(file)

	// Grab the header
	fileScanner.Scan()
	csvRef := csv.NewReader(bytes.NewReader(fileScanner.Bytes()))
	headerRef, _ := csvRef.Read()

	if options.MaxColumns != -1 && (len(headerRef) > options.MaxColumns || len(headerRef) == 0) {
		// Truncate the file and create a .ignore file
		reason := fmt.Sprintf("Bad number of (%d) columns: %s\n", len(headerRef), filename)
		markDead(filename, reason)
		return false
	}

	lineCount := 0
	for fileScanner.Scan() {
		lineCount++
		if lineCount >= options.MinLines {
			break
		}
	}

	file.Close()

	if lineCount < options.MinLines {
		// Truncate the file and create a .ignore file
		reason := fmt.Sprintf("Too few  (%d) lines : %s\n", lineCount, filename)
		markDead(filename, reason)
		return false
	} else {
		fmt.Printf(">>>> >%d lines, %d columns: %s\n", lineCount, len(headerRef), filename)
	}

	return true
}

func markDead(filename string, reason string) {
	// f, _ := os.Create(filename)
	// f.Close()

	f, _ := os.Create(filename + ".ignore")
	f.WriteString(reason)
	f.Close()
}
