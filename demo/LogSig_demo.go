package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type PlaceHolder struct{}

type Parser interface {
	Init()
	Parse()
	LoadLog()
}

type WordPair struct {
	a string
	b string
}

// type LogParser struct {
// }

type LogSig struct {
	inputDir   string
	outputDir  string
	logFile    string
	logFormat  string
	regexList  []string
	clusterNum int
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

func (parser *LogSig) Init(inputDir string, outputDir string, logFile string, logFormat string, regexList []string, clusterNum int) {
	parser.inputDir = inputDir
	parser.outputDir = outputDir
	parser.logFile = logFile
	parser.logFormat = logFormat
	parser.regexList = regexList
	parser.clusterNum = clusterNum
}

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

// TODO: wrap this function for dataFrame
func (parser *LogSig) GetLogContent(dataFrame map[string][]string) [][]string {
	wordSeqs := [][]string{}
	for _, content := range dataFrame["Content"] {
		// TODO: add regex processing
		content := strings.TrimRight(content, " ")
		wordSeq := strings.Split(content, " ")
		wordSeqs = append(wordSeqs, wordSeq)
	}
	return wordSeqs
}

func (parser *LogSig) LoadLog() ([]string, map[string][]string) {
	headers, regex := GenerateLogFormat(parser.logFormat)
	logFilePath := path.Join(parser.inputDir, parser.logFile)
	fmt.Println("Try to open logFile:", logFilePath)
	fmt.Println(regex, headers)
	dataFrame := LogToDataFrame(logFilePath, regex, headers, parser.logFormat)
	return headers, dataFrame
}

func WordSeqToPairs(wordSeqs [][]string) map[int][]WordPair {
	wordPairs := map[int][]WordPair{}
	for logId, wordSeq := range wordSeqs {
		wordPairs[logId] = []WordPair{}
		for i := 0; i < len(wordSeq); i++ {
			for j := i + 1; j < len(wordSeq); j++ {
				pair := WordPair{wordSeq[i], wordSeq[j]}
				wordPairs[logId] = append(wordPairs[logId], pair)
			}
		}
	}
	return wordPairs
}

func potentialDelta(currentCluster int, newCluster int, logId int, wordPairs map[int][]WordPair,
	pairRecord map[int]map[WordPair]map[int]struct{}, clusters map[int]map[int]struct{}) float64 {

	var delta float64 = 0.0
	for _, pair := range wordPairs[logId] {
		if _, ok := pairRecord[newCluster][pair]; ok {
			newClusterLength := len(clusters[newCluster]) + 1 // +1 in case the empty cluster
			currentClusterLength := len(clusters[currentCluster])
			delta += math.Pow(float64(len(pairRecord[newCluster][pair])/newClusterLength), 2) -
				math.Pow(float64(len(pairRecord[currentCluster][pair])/currentClusterLength), 2)
		} else {
			delta -= math.Pow(float64(len(pairRecord[currentCluster][pair])), 2)
		}
	}
	return delta
}

func (parser *LogSig) findClusterWithMaxPotential(currentCluster int, logId int, wordPairs map[int][]WordPair,
	pairRecord map[int]map[WordPair]map[int]struct{}, clusters map[int]map[int]struct{}) int {
	var maxDelta float64 = 0.0
	optimalCluster := currentCluster
	for newCluster := 0; newCluster < parser.clusterNum; newCluster++ {
		delta := potentialDelta(currentCluster, newCluster, logId, wordPairs, pairRecord, clusters)
		if delta > maxDelta {
			optimalCluster = newCluster
			maxDelta = delta
		}
	}
	return optimalCluster
}

func (parser *LogSig) LogCluster(wordPairs map[int][]WordPair) (map[int]int, map[int]map[int]struct{}) {
	clusters := map[int]map[int]struct{}{}
	clusterRecord := map[int]int{}
	pairRecord := map[int]map[WordPair]map[int]struct{}{}
	// initialize clusters
	for i := 0; i < parser.clusterNum; i++ {
		clusters[i] = map[int]struct{}{}
		pairRecord[i] = map[WordPair]map[int]struct{}{}
	}
	// randomly assign message into a cluster, also count word pairs
	for logId, pairs := range wordPairs {
		// assign cluster
		clusterId := rand.Intn(parser.clusterNum)
		clusterRecord[logId] = clusterId
		// count pairs
		for _, pair := range pairs {
			if _, ok := pairRecord[clusterId][pair]; ok {
				pairRecord[clusterId][pair][logId] = struct{}{}
			} else {
				pairRecord[clusterId][pair] = map[int]struct{}{logId: struct{}{}}
			}
		}
	}
	// group clusters by clusterId
	for logId, clusterId := range clusterRecord {
		if _, ok := clusters[clusterId]; ok {
			clusters[clusterId][logId] = struct{}{}
		} else {
			clusters[clusterId] = map[int]struct{}{logId: struct{}{}}
		}
	}
	// fmt.Println(clusterRecord)
	// local search
	changed := true
	for changed {
		changed = false
		for logId, pairs := range wordPairs {
			currentCluster := clusterRecord[logId]
			// search the new cluster that wold maximum the potential
			alterCluster := parser.findClusterWithMaxPotential(currentCluster, logId, wordPairs, pairRecord, clusters)
			if alterCluster == currentCluster {
				continue
			}
			// update cluster record and word pair record
			clusterRecord[logId] = alterCluster
			for _, pair := range pairs {
				if _, ok := pairRecord[alterCluster][pair]; ok {
					pairRecord[alterCluster][pair][logId] = struct{}{}
				} else {
					pairRecord[alterCluster][pair] = map[int]struct{}{logId: struct{}{}}
				}
				delete(pairRecord[currentCluster][pair], logId)
			}
			// update cluster
			delete(clusters[currentCluster], logId)
			clusters[alterCluster][logId] = struct{}{}
			changed = true
		}
	}
	// fmt.Println("ClusterRecord", clusterRecord)
	return clusterRecord, clusters
}

func (parser *LogSig) PatternExtract(clusterRecord map[int]int, wordSeqs [][]string, clusters map[int]map[int]struct{}) map[int]int {
	// group cluster by clusterId
	patterns := map[int]int{}
	for clusterId, logIds := range clusters {
		wordCount := map[string]int{}
		candidateWords := map[string]struct{}{}
		// calcuate word freqency for each group
		// select candidates words with frequency more than half of number of logs
		for logId, _ := range logIds {
			for _, word := range wordSeqs[logId] {
				if _, ok := wordCount[word]; ok {
					wordCount[word] += 1
				} else {
					wordCount[word] = 1
				}
				if wordCount[word] >= len(logIds)/2 {
					candidateWords[word] = struct{}{}
				}
			}
		}
		// choose the log with most candidate words as pattern
		patternId := 0
		maxMatches := 0
		for logId, _ := range logIds {
			matches := 0
			for _, word := range wordSeqs[logId] {
				if _, ok := candidateWords[word]; ok {
					matches += 1
				}
			}
			if matches > maxMatches {
				maxMatches = matches
				patternId = logId
			}
		}
		patterns[clusterId] = patternId
	}
	return patterns
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

func (parser *LogSig) WriteResultToFile(headers []string, wordSeqs [][]string, clusters map[int]map[int]struct{},
	patterns map[int]int, dataFrame map[string][]string, clusterRecord map[int]int) {

	// write log templates
	logTemplates := [][]string{[]string{"EventId", "EventTemplate"}}
	// iterate based on cluster index to keep order
	for i := 0; i < len(patterns); i++ {
		logId := patterns[i]
		patternSeq := wordSeqs[logId]
		pattern := strings.Join(patternSeq[:], " ")
		eventId := "Event" + strconv.Itoa(i+1)
		logTemplates = append(logTemplates, []string{eventId, pattern})
	}

	ExportCsvFile(parser.outputDir, parser.logFile+"_templates.csv", logTemplates)
	// write log structured
	logStructured := make([][]string, len(wordSeqs))
	for i, _ := range logStructured {
		lineId := "LineId" + strconv.Itoa(i+1)
		logStructured[i] = []string{lineId}
	}
	for _, header := range headers {
		records := dataFrame[header]
		// fmt.Println(header)
		for j, record := range records {
			logStructured[j] = append(logStructured[j], record)
		}
		for j, _ := range records {
			clusterId := clusterRecord[j]
			template := logTemplates[clusterId+1] // +1 here due to header in logTemplates
			eventId := template[0]
			pattern := template[1]
			logStructured[j] = append(logStructured[j], []string{eventId, pattern}...)
			// fmt.Println(logStructured[j])
		}
	}
	headers = append([]string{"LineId"}, headers...)
	headers = append(headers, []string{"EventId","EventTemplate"}...)
	logStructured = append([][]string{headers}, logStructured...)
	ExportCsvFile(parser.outputDir, parser.logFile+"_structured.csv", logStructured)
}

func (parser *LogSig) Parse() {
	startTime := time.Now()
	headers, dataFrame := parser.LoadLog()
	wordSeqs := parser.GetLogContent(dataFrame)
	// fmt.Println(wordSeqs)
	wordPairs := WordSeqToPairs(wordSeqs)
	clusterRecord, clusters := parser.LogCluster(wordPairs)
	// fmt.Println(clusterRecord)
	patterns := parser.PatternExtract(clusterRecord, wordSeqs, clusters)
	// fmt.Println(len(clusters))
	for _, logId := range patterns {
		fmt.Println(wordSeqs[logId])
	}
	parser.WriteResultToFile(headers, wordSeqs, clusters, patterns, dataFrame, clusterRecord)
	endTime := time.Now()
	fmt.Println("Parsing Done. Time taken: ", endTime.Sub(startTime))
}

func main() {
	// TODO:
	var parser LogSig
	inputDir := "../logs/HDFS/"
	outputDir := "LogSig_result/"
	logFile := "HDFS_2k.log"
	logFormat := "<Date> <Time> <Pid> <Level> <Component>: <Content>"
	regexList := []string{}
	groupNum := 14
	parser.Init(inputDir, outputDir, logFile, logFormat, regexList, groupNum)
	parser.Parse()
}
