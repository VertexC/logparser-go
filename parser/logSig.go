package parser

import (
	"math"
	"math/rand"
	// "time"
	"strings"
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
	model.parser.Init(inputDir, outputDir, logFile, logFormat, regexList)
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
	currentClusterLength := len(clusters[currentCluster])
	for _, pair := range wordPairs[logId] {
		newRecord := 1
		if _, ok := pairRecord[newCluster][pair]; ok {
			newRecord += len(pairRecord[newCluster][pair])
		}
		newClusterLength := len(clusters[newCluster]) + 1 // +1 in case the empty cluster
		pNew := float64(newRecord) / float64(newClusterLength)
		pCurrent := float64(len(pairRecord[currentCluster][pair])) / float64(currentClusterLength)
		delta += math.Pow(pNew, 2) - math.Pow(pCurrent, 2)
	}
	// omit factor 3 here
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
				pairRecord[clusterId][pair] = map[int]struct{}{logId: {}}
			}
		}
	}
	// group clusters by clusterId
	for logId, clusterId := range clusterRecord {
		if _, ok := clusters[clusterId]; ok {
			clusters[clusterId][logId] = struct{}{}
		} else {
			clusters[clusterId] = map[int]struct{}{logId: {}}
		}
	}
	// fmt.Println(clusterRecord)
	// local search
	changed := true
	rounds := 0
	for changed {
		changed = false
		for logId, pairs := range wordPairs {
			currentCluster := clusterRecord[logId]
			// search the new cluster that word maximum the potential
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
					pairRecord[alterCluster][pair] = map[int]struct{}{logId: {}}
				}
				delete(pairRecord[currentCluster][pair], logId)
			}
			// update cluster
			delete(clusters[currentCluster], logId)
			clusters[alterCluster][logId] = struct{}{}
			changed = true
		}
		rounds += 1
	}
	// fmt.Println("ClusterRecord", clusterRecord)
	return clusterRecord, clusters
}

func (model *LogSig) PatternExtract(clusterRecord map[int]int, wordSeqs [][]string, clusters map[int]map[int]struct{}) map[int]string {
	// group cluster by clusterId
	patterns := map[int]string{}
	for clusterId, logIds := range clusters {
		wordCount := map[string]int{}
		candidateWords := map[string]struct{}{}
		// calcuate word freqency for each group
		// select candidates words with frequency more than half of number of logs
		for logId := range logIds {
			for _, word := range wordSeqs[logId] {
				if _, ok := wordCount[word]; ok {
					wordCount[word] += 1
				} else {
					wordCount[word] = 1
				}
				if float32(wordCount[word]) >= float32(len(logIds))/2.0 {
					candidateWords[word] = struct{}{}
				}
			}
		}
		// scan each log in the cluster
		// extract candidate word in that log
		// for a candidate pattern
		// select the candidate pattern with the most occurence
		maxOccurence := 0
		pattern := ""
		occurenceRecord := map[string]int{}
		for logId := range logIds {
			// construct candidate pattern
			candidatePatternSeq := []string{}
			for _, word := range wordSeqs[logId] {
				if _, ok := candidateWords[word]; ok {
					candidatePatternSeq = append(candidatePatternSeq, word)
				}
			}
			candidatePattern := strings.Join(candidatePatternSeq, " ")
			if _, ok := occurenceRecord[candidatePattern]; ok {
				occurenceRecord[candidatePattern] += 1
			} else {
				occurenceRecord[candidatePattern] = 1
			}
			if occurenceRecord[candidatePattern] > maxOccurence {
				maxOccurence = occurenceRecord[candidatePattern]
				pattern = candidatePattern
			}
		}
		patterns[clusterId] = pattern
	}
	return patterns
}

func (model *LogSig) Parse() {
	headers, dataFrame := model.parser.LoadLog()
	wordSeqs := model.parser.GetLogContent(dataFrame)
	wordPairs := WordSeqToPairs(wordSeqs)
	clusterRecord, clusters := model.LogCluster(wordPairs)
	clusterSize := map[int]int{}
	for clusterId, cluster := range clusters {
		clusterSize[clusterId] = len(cluster)
	}
	patterns := model.PatternExtract(clusterRecord, wordSeqs, clusters)
	model.parser.WriteResultToFile(headers, dataFrame, wordSeqs, patterns, clusterRecord, clusterSize, false)
}
