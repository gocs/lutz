package lutz

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"strings"
	"sync"
)

// GetLookupTable returns a lookup table of the timezone data
func GetLookupTable(r io.Reader, w io.Writer) error {

	var wg sync.WaitGroup
	// extract the timezone file from the input file
	in := extr(r)

	// process the timezone file
	out := proc(in, &wg)

	// sort the timezone file
	sorted := sort(out)

	// write the sorted timezone file to the output file
	for _, v := range sorted {
		w.Write([]byte(v + "\n"))
	}
	wg.Wait()
	return nil
}

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

// extr extracts the timezone file from the input file line by line
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

// proc processes the timezone file to a formatted string (<continent>/<city>/<subcategory> <hour> <minutes>)
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

// sort sorts the timezone file upon insertion
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
