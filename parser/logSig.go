package parser

import (
	"math"
	"math/rand"
	"strconv"
	"strings"
	// "time"
)

type WordPair struct {
	a string
	b string
}

type LogSig struct {
	parser     Parser
	clusterNum int
}

func (model *LogSig) Init(inputDir string, outputDir string, logFile string, logFormat string, regexList []string, clusterNum int) {
	model.parser.Init(inputDir, outputDir, logFile, logFormat, regexList, clusterNum)
	model.clusterNum = clusterNum
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

func (model *LogSig) findClusterWithMaxPotential(currentCluster int, logId int, wordPairs map[int][]WordPair,
	pairRecord map[int]map[WordPair]map[int]struct{}, clusters map[int]map[int]struct{}) int {
	var maxDelta float64 = 0.0
	optimalCluster := currentCluster
	for newCluster := 0; newCluster < model.clusterNum; newCluster++ {
		delta := potentialDelta(currentCluster, newCluster, logId, wordPairs, pairRecord, clusters)
		if delta > maxDelta {
			optimalCluster = newCluster
			maxDelta = delta
		}
	}
	return optimalCluster
}

func (model *LogSig) LogCluster(wordPairs map[int][]WordPair) (map[int]int, map[int]map[int]struct{}) {
	clusters := map[int]map[int]struct{}{}
	clusterRecord := map[int]int{}
	pairRecord := map[int]map[WordPair]map[int]struct{}{}
	// initialize clusters
	for i := 0; i < model.clusterNum; i++ {
		clusters[i] = map[int]struct{}{}
		pairRecord[i] = map[WordPair]map[int]struct{}{}
	}
	// randomly assign message into a cluster, also count word pairs
	for logId, pairs := range wordPairs {
		// assign cluster
		clusterId := rand.Intn(model.clusterNum)
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
			alterCluster := model.findClusterWithMaxPotential(currentCluster, logId, wordPairs, pairRecord, clusters)
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

func (model *LogSig) PatternExtract(clusterRecord map[int]int, wordSeqs [][]string, clusters map[int]map[int]struct{}) map[int]int {
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

func (model *LogSig) WriteResultToFile(headers []string, wordSeqs [][]string, clusters map[int]map[int]struct{},
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

	ExportCsvFile(model.parser.outputDir, model.parser.logFile+"_templates.csv", logTemplates)
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
	headers = append(headers, []string{"EventId", "EventTemplate"}...)
	logStructured = append([][]string{headers}, logStructured...)
	ExportCsvFile(model.parser.outputDir, model.parser.logFile+"_structured.csv", logStructured)
}

func (model *LogSig) Parse() {
	// startTime := time.Now()
	headers, dataFrame := model.parser.LoadLog()
	wordSeqs := model.parser.GetLogContent(dataFrame)
	// fmt.Println(wordSeqs)
	wordPairs := WordSeqToPairs(wordSeqs)
	clusterRecord, clusters := model.LogCluster(wordPairs)
	// fmt.Println(clusterRecord)
	patterns := model.PatternExtract(clusterRecord, wordSeqs, clusters)
	// fmt.Println(len(clusters))
	// for _, logId := range patterns {
	// 	fmt.Println(wordSeqs[logId])
	// }
	model.WriteResultToFile(headers, wordSeqs, clusters, patterns, dataFrame, clusterRecord)
	// endTime := time.Now()
	// fmt.Println("Parsing Done. Time taken: ", endTime.Sub(startTime))
}
