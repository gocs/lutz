package main

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
)

type Key string

const (
	Africa       Key = "africa"
	Antarctica   Key = "antarctica"
	Asia         Key = "asia"
	Europe       Key = "europe"
	Australasia  Key = "australasia"
	NorthAmerica Key = "northamerica"
	SouthAmerica Key = "southamerica"
)

var Continents = map[Key]string{
	Africa:       "africa",
	Antarctica:   "antarctica",
	Asia:         "asia",
	Europe:       "europe",
	Australasia:  "australasia",
	NorthAmerica: "northamerica",
	SouthAmerica: "southamerica",
}

const (
	timezoneFileURL = "https://data.iana.org/time-zones/releases"
	timezoneFileIn  = "tzdata2025c.tar.gz"
	timezoneFileOut = "tz"
)

// dl download the timezone file from the url to the writer
func dl(ctx context.Context, w io.Writer) error {
	urlPath, err := url.JoinPath(timezoneFileURL, timezoneFileIn)
	if err != nil {
		return fmt.Errorf("error joining url: %w", err)
	}
	request, err := http.NewRequestWithContext(ctx, "GET", urlPath, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return fmt.Errorf("error doing request: %w", err)
	}
	defer response.Body.Close()
	_, err = io.Copy(w, response.Body)
	if err != nil {
		return fmt.Errorf("error copying response body: %w", err)
	}
	return nil
}

func extr(r io.Reader) <-chan string {
	out := make(chan string)
	go func() {
		defer close(out)

		gr, err := gzip.NewReader(r)
		if err != nil {
			log.Fatalf("error creating gzip reader: %v", err)
		}
		defer gr.Close()

		tr := tar.NewReader(gr)
		for {
			h, err := tr.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalf("error traversing tar file: %v", err)
			}
			if _, ok := Continents[Key(h.Name)]; !ok {
				continue
			}

			s := bufio.NewScanner(tr)
			for s.Scan() {
				out <- s.Text()
			}
			if err := s.Err(); err != nil {
				log.Fatalf("error scanning file: %v", err)
			}
		}
	}()
	return out
}

func proc(in <-chan string, wg *sync.WaitGroup) <-chan string {
	out := make(chan string)
	wg.Go(func() {
		defer close(out)

		currentZone := ""
		latestOffset := ""
		fmtOffset := ""

		for s := range in {

			// trim space
			line := strings.TrimSpace(s)
			if line == "" {
				continue
			}

			// skip comments
			if strings.HasPrefix(line, "#") {
				continue
			}

			// skip rules
			if strings.HasPrefix(line, "Rule") {
				continue
			}

			if strings.HasPrefix(line, "Zone") {
				// Save previous zone if we have one
				if currentZone != "" && latestOffset != "" {
					fmtOffset = formatOffset(latestOffset)
					formatted := fmt.Sprintf("%s %s", currentZone, fmtOffset)
					out <- formatted
				}

				// Extract zone name (second column)
				zoneName := strings.Fields(line)[1]
				zoneName = strings.Replace(zoneName, " ", "\t", 1)
				currentZone = zoneName
			} else {
				// Continuation line - extract offset (first non-whitespace part)
				offset := strings.TrimSpace(strings.Fields(line)[0])
				latestOffset = offset
			}
		}

		// Handle last zone if we have one
		if currentZone != "" && latestOffset != "" {
			fmtOffset = formatOffset(latestOffset)
			formatted := fmt.Sprintf("%s %s", currentZone, fmtOffset)
			out <- formatted
		}
	})
	return out
}

func sort(in <-chan string) []string {
	sorted := []string{}
	for s := range in {
		if len(sorted) == 0 {
			sorted = append(sorted, s)
			continue
		}

		// insert s into sorted
		for _, v := range sorted {
			// b > a = [a, b]
			if s > v {
				sorted = append(sorted, s)
				break
			}
		}
	}
	return sorted
}

// formatOffset Convert offset from 'h:mm' format to 'h mm' format, trimming leading zero from hour
func formatOffset(offset string) string {
	fields := strings.Split(offset, ":")
	return fmt.Sprintf("%s %s", fields[0], fields[1])
}

func main() {
	ctx := context.Background()

	// open the input file for reading from the source
	fileIn := bytes.NewBuffer(nil)
	// fileIn, err := os.Open(timezoneFileIn)
	// if err != nil {
	// 	log.Fatalf("error opening file: %v", err)
	// }
	// defer fileIn.Close()

	// open the output file for writing to the destination
	fileOut, err := os.Create(timezoneFileOut)
	if err != nil {
		log.Fatalf("error creating file: %v", err)
	}
	defer fileOut.Close()

	if err := dl(ctx, fileIn); err != nil {
		log.Fatalf("error downloading file: %v", err)
	}

	var wg sync.WaitGroup

	// extract the timezone file from the input file
	in := extr(fileIn)

	// process the timezone file
	out := proc(in, &wg)

	// sort the timezone file
	sorted := sort(out)

	// write the sorted timezone file to the output file
	for _, v := range sorted {
		fileOut.WriteString(v + "\n")
	}

	wg.Wait()
}
