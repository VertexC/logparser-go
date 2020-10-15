package parser

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path"
	"regexp"
	"strings"
)

func GenerateLogFormat(format string) ([]string, string) {
	var headers []string
	var regex string
	matches := regexp.MustCompile(`(<[^<>]+>)`).FindAllStringIndex(format, -1)
	// fmt.Println(format)
	splitters := []string{}
	// fmt.Println(matches)
	var lastMatch []int
	for i, match := range matches {
		if i != 0 {
			gap := []int{lastMatch[1], match[0]}
			if gap[0] != gap[1] {
				splitters = append(splitters, format[gap[0]:gap[1]])
			}
			lastMatch = match
		}
		lastMatch = match
		splitters = append(splitters, format[match[0]:match[1]])
	}
	for i, s := range splitters {
		if i%2 == 1 {
			re := regexp.MustCompile(` +`)
			s = re.ReplaceAllString(s, `\s+`)
			regex += s
		} else {
			header := strings.Trim(s, "<")
			header = strings.Trim(header, ">")
			headers = append(headers, header)
			regex += fmt.Sprintf(`(?P<%s>.*?)`, header)
		}
	}
	regex = `^` + regex + `$`
	return headers, regex
}

func ReadLines(filePath string) []string {
	var lines []string
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return lines
}

func getParams(re *regexp.Regexp, s string) map[string]string {
	match := re.FindStringSubmatch(s)
	paramsMap := make(map[string]string)
	for i, name := range re.SubexpNames() {
		if i > 0 && i <= len(match) {
			paramsMap[name] = match[i]
		}
	}
	return paramsMap
}

func LogToDataFrame(logFilePath string, regex string, headers []string, logFormat string) map[string][]string {
	lines := ReadLines(logFilePath)
	re := regexp.MustCompile(regex)
	dataFrame := map[string][]string{}
	for _, line := range lines {
		// fmt.Println("[logToDataFrame]", i, line)
		params := getParams(re, line)
		// fmt.Println(params)
		for header, data := range params {
			if _, ok := dataFrame[header]; ok {
				dataFrame[header] = append(dataFrame[header], data)
			} else {
				dataFrame[header] = []string{data}
			}
		}
	}
	// fmt.Println(dataFrame)
	// fmt.Println(re.SubexpNames())
	return dataFrame
}

func ExportCsvFile(outputDir string, fileName string, records [][]string) error {
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		log.Fatalln("error when create dir:", err)
		return err
	}

	filePath := path.Join(outputDir, fileName)
	file, err := os.Create(filePath)
	if err != nil {
		log.Fatalln("error create file:", err)
		return err
	}
	defer file.Close()

	w := csv.NewWriter(file)
	w.WriteAll(records)

	if err := w.Error(); err != nil {
		log.Fatalln("error writing csv:", err)
	}
	return nil
}

// https://stackoverflow.com/questions/24999079/reading-csv-file-in-go
func ReadCsvFile(filePath string) [][]string {
	f, err := os.Open(filePath)
	if err != nil {
		log.Fatal("Unable to read input file "+filePath, err)
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	records, err := csvReader.ReadAll()
	if err != nil {
		log.Fatal("Unable to parse file as CSV for"+filePath, err)
	}
	return records
}
