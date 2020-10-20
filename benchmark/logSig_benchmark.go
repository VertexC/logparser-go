package benchmark

import (
	"fmt"
	"github.com/VertexC/logparser-go/parser"
	"path"
	"time"
)

type LogSigSetting struct {
	inputDir  string
	logFile   string
	logFormat string
	regexList []string
	groupNum  int
}

func LogSigBenchmark() {
	outputDir := "benchmark/LogSig_result"
	benchmarkSettings := map[string]LogSigSetting{
		"HDFS": {
			"logs/HDFS",
			"HDFS_2k.log",
			"<Date> <Time> <Pid> <Level> <Component>: <Content>",
			[]string{`blk_-?\d+`, `(\d+\.){3}\d+(:\d+)?`},
			15},
		"Hadoop": {
			"logs/Hadoop",
			"Hadoop_2k.log",
			`<Date> <Time> <Level> \[<Process>\] <Component>: <Content>`,
			[]string{`(\d+\.){3}\d+`},
			30},
	}

	for dataSetName, setting := range benchmarkSettings {
		fmt.Println("Start to Run LogSig on: ", dataSetName)
		startTime := time.Now()

		var model parser.LogSig
		model.Init(setting.inputDir, outputDir, setting.logFile, setting.logFormat,
			setting.regexList, setting.groupNum)
		model.Parse()
		parser.Evaluate(path.Join(outputDir, setting.logFile+"_structured.csv"),
			path.Join(setting.inputDir, setting.logFile+"_structured.csv"))
		endTime := time.Now()
		fmt.Println("Time duration: ", endTime.Sub(startTime))
	}
}
