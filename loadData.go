package sego

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
)

// 从文件中载入词典
//
// 可以载入多个词典文件，文件名用","分隔，排在前面的词典优先载入分词，比如
// 	"用户词典.txt,通用词典.txt"
// 当一个分词既出现在用户词典也出现在通用词典中，则优先使用用户词典。
//
// 词典的格式为（每个分词一行）：
//	分词文本 频率 词性
func (seg *Segmenter) LoadDictionary(files string) {
	seg.dict = NewDictionary()
	for _, file := range strings.Split(files, ",") {
		log.Printf("载入sego词典 %s", file)
		dictFile, err := os.Open(file)
		defer dictFile.Close()
		if err != nil {
			log.Fatalf("无法载入字典文件 \"%s\" \n", file)
		}

		reader := bufio.NewReader(dictFile)
		var text string
		var freqText string
		var frequency int
		var pos string

		// 逐行读入分词
		for {
			size, _ := fmt.Fscanln(reader, &text, &freqText, &pos)

			if size == 0 {
				// 文件结束
				break
			} else if size < 2 {
				// 无效行
				continue
			} else if size == 2 {
				// 没有词性标注时设为空字符串
				pos = ""
			}

			// 解析词频
			var err error
			frequency, err = strconv.Atoi(freqText)
			if err != nil {
				continue
			}

			// 过滤频率太小的词
			if frequency < minTokenFrequency {
				continue
			}

			// 将分词添加到字典中
			words := splitTextToWords([]byte(text))
			token := Token{text: words, frequency: frequency, pos: pos}
			seg.dict.addToken(token)
		}
	}

	// 计算每个分词的路径值，路径值含义见Token结构体的注释
	logTotalFrequency := float32(math.Log2(float64(seg.dict.totalFrequency)))
	for i := range seg.dict.tokens {
		token := &seg.dict.tokens[i]
		token.distance = logTotalFrequency - float32(math.Log2(float64(token.frequency)))
	}

	// 对每个分词进行细致划分，用于搜索引擎模式，该模式用法见Token结构体的注释。
	for i := range seg.dict.tokens {
		token := &seg.dict.tokens[i]
		segments := seg.cutJump(token.text, true)

		// 计算需要添加的子分词数目
		numTokensToAdd := 0
		for iToken := 0; iToken < len(segments); iToken++ {
			if len(segments[iToken].token.text) > 0 {
				numTokensToAdd++
			}
		}
		token.segments = make([]*Segment, numTokensToAdd)

		// 添加子分词
		iSegmentsToAdd := 0
		for iToken := 0; iToken < len(segments); iToken++ {
			if len(segments[iToken].token.text) > 0 {
				token.segments[iSegmentsToAdd] = &segments[iToken]
				iSegmentsToAdd++
			}
		}
	}

	log.Println("sego词典载入完毕")
}

// LoadDictionaryFromDB 从数据库导入词库
func (seg *Segmenter) LoadDictionaryFromDB(infos []map[string]string) {
	seg.dict = NewDictionary()
	for _, v := range infos {
		// 解析词频
		frequency, err := strconv.Atoi(v["freqText"])
		if err != nil {
			continue
		}

		// 过滤频率太小的词
		if frequency < minTokenFrequency {
			continue
		}

		// 将分词添加到字典中
		words := splitTextToWords([]byte(v["text"]))

		token := Token{text: words, frequency: frequency, pos: v["pos"]}
		seg.dict.addToken(token)
	}

	// 计算每个分词的路径值，路径值含义见Token结构体的注释
	logTotalFrequency := float32(math.Log2(float64(seg.dict.totalFrequency)))
	for i := range seg.dict.tokens {
		token := &seg.dict.tokens[i]
		token.distance = logTotalFrequency - float32(math.Log2(float64(token.frequency)))
	}

	// 对每个分词进行细致划分，用于搜索引擎模式，该模式用法见Token结构体的注释。
	for i := range seg.dict.tokens {
		token := &seg.dict.tokens[i]
		segments := seg.cutJump(token.text, true)

		// 计算需要添加的子分词数目
		numTokensToAdd := 0
		for iToken := 0; iToken < len(segments); iToken++ {
			if len(segments[iToken].token.text) > 0 {
				numTokensToAdd++
			}
		}
		token.segments = make([]*Segment, numTokensToAdd)

		// 添加子分词
		iSegmentsToAdd := 0
		for iToken := 0; iToken < len(segments); iToken++ {
			if len(segments[iToken].token.text) > 0 {
				token.segments[iSegmentsToAdd] = &segments[iToken]
				iSegmentsToAdd++
			}
		}
	}

	log.Println("sego词典载入完毕")
}
