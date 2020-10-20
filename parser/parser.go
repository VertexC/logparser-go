package parser

import (
	"path"
	"regexp"
	"strconv"
	"strings"
	// "fmt"
)

type Parser struct {
	inputDir  string
	outputDir string
	logFile   string
	logFormat string
	regexList []string
}

func (parser *Parser) Init(inputDir string, outputDir string, logFile string, logFormat string, regexList []string) {
	parser.inputDir = inputDir
	parser.outputDir = outputDir
	parser.logFile = logFile
	parser.logFormat = logFormat
	parser.regexList = regexList
}

// TODO: wrap this function for dataFrame
func (parser *Parser) GetLogContent(dataFrame map[string][]string) [][]string {
	wordSeqs := [][]string{}
	for _, content := range dataFrame["Content"] {
		// TODO: add regex processing
		content := strings.TrimRight(content, " ")
		for _, regex := range parser.regexList {
			re := regexp.MustCompile(regex)
			content = re.ReplaceAllString(content, "<*>")
		}
		wordSeq := strings.Split(content, " ")
		wordSeqs = append(wordSeqs, wordSeq)
	}
	return wordSeqs
}

func (parser *Parser) LoadLog() ([]string, map[string][]string) {
	headers, regex := GenerateLogFormat(parser.logFormat)
	logFilePath := path.Join(parser.inputDir, parser.logFile)
	// fmt.Println(regex, headers)
	dataFrame := LogToDataFrame(logFilePath, regex, headers, parser.logFormat)
	return headers, dataFrame
}

func ExtractParams(content string, pattern string) []string {
	// deal with invalid regex symbol
	regex := regexp.MustCompile(`([^A-Za-z0-9])`).ReplaceAllString(pattern, `\$1`)
	// deal with white spaces
	regex = regexp.MustCompile(`\ +`).ReplaceAllString(regex, `\s+`)
	regex = regexp.MustCompile(`\<\*\>`).ReplaceAllString(regex, `(.*)?`)
	// fmt.Println(regex)
	re := regexp.MustCompile(`^` + regex + `$`)
	params := re.FindStringSubmatch(content)
	if len(params) >= 1 {
		return params[1:]
	}
	return []string{}
}

func (parser *Parser) WriteResultToFile(headers []string, dataFrame map[string][]string, wordSeqs [][]string,
	patterns map[int]string, clusterRecord map[int]int, clusterSize map[int]int, setParameter bool) {

	// write log templates
	logTemplates := [][]string{{"EventId", "EventTemplate", "Occurrences"}}
	// iterate based on cluster index to keep order
	for i := 0; i < len(patterns); i++ {
		pattern := patterns[i]
		eventId := "Event" + strconv.Itoa(i+1)
		occurrence := strconv.Itoa(clusterSize[i])
		logTemplates = append(logTemplates, []string{eventId, pattern, occurrence})
	}

	ExportCsvFile(parser.outputDir, parser.logFile+"_templates.csv", logTemplates)
	// write log structured
	logStructured := make([][]string, len(wordSeqs))
	for i := range logStructured {
		lineId := "LineId" + strconv.Itoa(i+1)
		logStructured[i] = []string{lineId}
	}
	for _, header := range headers {
		records := dataFrame[header]
		// fmt.Println(header)
		for j, record := range records {
			logStructured[j] = append(logStructured[j], record)
		}

	}
	for j := range wordSeqs {
		clusterId := clusterRecord[j]
		template := logTemplates[clusterId+1] // +1 here due to header in logTemplates
		eventId := template[0]
		pattern := template[1]
		logStructured[j] = append(logStructured[j], []string{eventId, pattern}...)
		if setParameter {
			params := ExtractParams(dataFrame["Content"][j], pattern)
			logStructured[j] = append(logStructured[j], strings.Join(params, ";"))
		}
	}
	headers = append([]string{"LineId"}, headers...)
	headers = append(headers, []string{"EventId", "EventTemplate"}...)
	if setParameter {
		headers = append(headers, "ParameterList")
	}
	logStructured = append([][]string{headers}, logStructured...)
	ExportCsvFile(parser.outputDir, parser.logFile+"_structured.csv", logStructured)
}
