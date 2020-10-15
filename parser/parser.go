package parser

import (
	"path"
	"strings"
)

type Parser struct {
	inputDir  string
	outputDir string
	logFile   string
	logFormat string
	regexList []string
}

func (parser *Parser) Init(inputDir string, outputDir string, logFile string, logFormat string, regexList []string, clusterNum int) {
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
		wordSeq := strings.Split(content, " ")
		wordSeqs = append(wordSeqs, wordSeq)
	}
	return wordSeqs
}

func (parser *Parser) LoadLog() ([]string, map[string][]string) {
	headers, regex := GenerateLogFormat(parser.logFormat)
	logFilePath := path.Join(parser.inputDir, parser.logFile)
	// fmt.Println("Try to open logFile:", logFilePath)
	// fmt.Println(regex, headers)
	dataFrame := LogToDataFrame(logFilePath, regex, headers, parser.logFormat)
	return headers, dataFrame
}
