package bench

import (
	"github.com/VertexC/logparser-go/parser"
	"testing"
)

func TestLogSig(t *testing.T) {
	inputDir := "../logs/HDFS/"
	outputDir := "LogSig_result/"
	logFile := "HDFS_2k.log"
	logFormat := "<Date> <Time> <Pid> <Level> <Component>: <Content>"
	regexList := []string{}
	groupNum := 14

	var model parser.LogSig
	model.Init(inputDir, outputDir, logFile, logFormat, regexList, groupNum)
	model.Parse()
}

func TestDrain(t *testing.T) {
	inputDir := "../logs/HDFS/"
	outputDir := "Drain_result/"
	logFile := "HDFS_2k.log"
	logFormat := "<Date> <Time> <Pid> <Level> <Component>: <Content>"
	regexList := []string{
		`blk_(|-)[0-9]+`,
		`(/|)([0-9]+\.){3}[0-9]+(:[0-9]+|)(:|)`, // IP
		// TODO: original python expr contains lookahead/lookbehind proceeding
		// r'(?<=[^A-Za-z0-9])(\-?\+?\d+)(?=[^A-Za-z0-9])|[0-9]+$'
		// if write as follows, the character like whitespace will be replace with digits together by <*>
		`[^A-Za-z0-9](\-?\+?\d+)[^A-Za-z0-9]|[0-9]+$`, // Numbers
	}
	similarityThreshold := float32(0.5)
	depth := 4

	var model parser.Drain
	model.Init(inputDir, outputDir, logFile, logFormat, regexList, similarityThreshold, depth)
	model.Parse()
}
