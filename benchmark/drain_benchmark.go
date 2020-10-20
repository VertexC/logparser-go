package benchmark

import (
	"fmt"
	"github.com/VertexC/logparser-go/parser"
	"path"
	"time"
)

type DrainSetting struct {
	inputDir  string
	logFile   string
	logFormat string
	regexList []string
	st        float32
	depth     int
}

func DrainBenchmark() {
	outputDir := "benchmark/Drain_result"
	benchmarkSettings := map[string]DrainSetting{
		"HDFS": {
			"logs/HDFS",
			"HDFS_2k.log",
			"<Date> <Time> <Pid> <Level> <Component>: <Content>",
			[]string{`blk_-?\d+`, `(\d+\.){3}\d+(:\d+)?`},
			0.5, 4},
		"Hadoop": {
			"logs/Hadoop",
			"Hadoop_2k.log",
			`<Date> <Time> <Level> \[<Process>\] <Component>: <Content>`,
			[]string{`(\d+\.){3}\d+`},
			0.5, 4},
	}

	for dataSetName, setting := range benchmarkSettings {
		fmt.Println("\nStart to Run Drain on: ", dataSetName)
		startTime := time.Now()

		var model parser.Drain
		model.Init(setting.inputDir, outputDir, setting.logFile, setting.logFormat,
			setting.regexList, setting.st, setting.depth)
		model.Parse()
		metric := parser.Evaluate(path.Join(outputDir, setting.logFile+"_structured.csv"),
			path.Join(setting.inputDir, setting.logFile+"_structured.csv"))
		fmt.Printf("metric: %+v\n", *metric)
		endTime := time.Now()
		fmt.Println("Time duration: ", endTime.Sub(startTime))
	}
}
