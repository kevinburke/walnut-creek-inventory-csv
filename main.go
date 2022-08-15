package main

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"flag"
	"log"
	"os"
	"regexp"
	"strings"
)

var zoningRx = regexp.MustCompile("^[A-Z]{1,2}-[A-Z0-9.]{1,4}$")

func main() {
	flag.Parse()
	data, err := os.ReadFile(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}
	data2, err := os.ReadFile(flag.Arg(1))
	if err != nil {
		log.Fatal(err)
	}
	parcelRx := regexp.MustCompile("^[0-9]{3}-[0-9]{3}-[0-9]{3}-[0-9]$")
	bs := bufio.NewScanner(bytes.NewReader(data))
	w := csv.NewWriter(os.Stdout)
	csv := make([][]string, 0)
	for bs.Scan() {
		line := bs.Text()
		csvfields := make([]string, 0)
		fields := strings.Fields(line)
		if fields[0] != "WALNUT" || fields[1] != "CREEK" {
			log.Fatalf("unexpected field 1: %q", fields)
		}
		csvfields = append(csvfields, "WALNUT CREEK")
		csvfields = append(csvfields, fields[2])
		// everything to the parcel number is the address
		match := false
		currentIdx := 3
		for i := 3; i < len(fields); i++ {
			if parcelRx.MatchString(fields[i]) {
				csvfields = append(csvfields, strings.Join(fields[3:i-1], " "))
				csvfields = append(csvfields, fields[i-1])
				csvfields = append(csvfields, fields[i])
				currentIdx = i + 1
				match = true
				break
			}
		}
		if !match {
			log.Fatalf("did not find a parcel number: %q", fields)
		}
		// if this field has a trailing comma it's a zoning designation,
		// otherwise it's a consolidated parcel
		if !strings.HasSuffix(fields[currentIdx], ",") {
			csvfields = append(csvfields, fields[currentIdx])
			currentIdx++
		} else {
			csvfields = append(csvfields, "")
		}
		match = false
		for i := currentIdx + 1; i < len(fields); i++ {
			if zoningRx.MatchString(fields[i]) {
				csvfields = append(csvfields, strings.Join(fields[currentIdx:i], " "))
				csvfields = append(csvfields, fields[i])
				match = true
				currentIdx = i + 1
				break
			}
		}
		if !match {
			log.Fatalf("did not find a zoning designation: %q", fields)
		}
		csvfields = append(csvfields, fields[currentIdx])
		currentIdx++
		csvfields = append(csvfields, fields[currentIdx])
		currentIdx++
		csvfields = append(csvfields, fields[currentIdx])
		currentIdx++
		csvfields = append(csvfields, strings.Join(fields[currentIdx:len(fields)], " "))
		csv = append(csv, csvfields)
	}
	if err := bs.Err(); err != nil {
		log.Fatal(err)
	}
	bs2 := bufio.NewScanner(bytes.NewReader(data2))
	csv2 := make([][]string, 0)
	for bs2.Scan() {
		line := bs2.Text()
		if line == "Site" {
			bs2.Scan()
			line = bs2.Text()
			if line != "Number" {
				log.Fatal("unexpected data: %q", line)
			}
		}
		count := 0
		csvOffset := len(csv2)
		for i := 0; true; i++ {
			bs2.Scan()
			line = bs2.Text()
			fields := strings.Fields(line)
			if len(fields) > 1 {
				if fields[0] != "Infrastructure" {
					log.Fatal("unexpected header line: %q", fields)
				}
				break
			}
			count++
			csv2 = append(csv2, []string{fields[0]})
		}
		for i := 0; i < count; i++ {
			bs2.Scan()
			line = bs2.Text()
			fields := strings.Fields(line)
			if fields[0] != "YES" || fields[1] != "-" || fields[2] != "Current" {
				log.Fatal("unexpected first field: %q", fields)
			}
			currentIdx := 3
			csv2[csvOffset+i] = append(csv2[csvOffset+i], "YES - Current")
			if fields[3] == "NO" && fields[4] == "-" && fields[5] == "Privately-" && fields[6] == "Owned" {
				csv2[csvOffset+i] = append(csv2[csvOffset+i], "No - Privately Owned")
				currentIdx = 7
			} else if fields[3] == "YES" && fields[4] == "-" && fields[5] == "Other" && fields[6] == "Publicly-" && fields[7] == "Owned" {
				csv2[csvOffset+i] = append(csv2[csvOffset+i], "Yes - Other Publicly-Owned")
				currentIdx = 8
			} else {
				log.Fatalf("unexpected owner field: %q", fields)
			}
			switch {
			case fields[currentIdx] == "Available":
				csv2[csvOffset+i] = append(csv2[csvOffset+i], "Available")
				currentIdx++
			case fields[currentIdx] == "Pending" && fields[currentIdx+1] == "Project":
				csv2[csvOffset+i] = append(csv2[csvOffset+i], "Pending Project")
				currentIdx += 2
			default:
				log.Fatalf("unexpected status: %q", fields)
			}
			rest := strings.Join(fields[currentIdx:], " ")
			switch {
			case strings.HasPrefix(rest, "Used in Prior Housing Element - Non-Vacant"):
				csv2[csvOffset+i] = append(csv2[csvOffset+i], "Used in Prior Housing Element - Non-Vacant")
				rest = rest[len("Used in Prior Housing Element - Non-Vacant "):]
			case strings.HasPrefix(rest, "Not Used in Prior Housing Element"):
				csv2[csvOffset+i] = append(csv2[csvOffset+i], "Not Used in Prior Housing Element")
				rest = rest[len("Not Used in Prior Housing Element "):]
			default:
				log.Fatalf("unexpected last cycle line: %q", rest)
			}
			fields = strings.Fields(rest)
			if len(fields) < 4 {
				log.Fatalf("too short line: %q", rest)
			}
			csv2[csvOffset+i] = append(csv2[csvOffset+i], []string{fields[0], fields[1], fields[2], fields[3]}...)
			currentIdx = 4
			csv2[csvOffset+i] = append(csv2[csvOffset+i], strings.Join(fields[4:], " "))
		}
	}
	if err := bs2.Err(); err != nil {
		log.Fatal(err)
	}
	for i := 0; i < len(csv2); i++ {
		csv[i] = append(csv[i], csv2[i]...)
	}
	w.WriteAll(csv)
	w.Flush()
}
