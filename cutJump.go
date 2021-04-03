package sego

func (seg *Segmenter) cutJump(text []Text, searchMode bool) []Segment {
	// 搜索模式下该分词已无继续划分可能的情况
	if searchMode && len(text) == 1 {
		return []Segment{}
	}

	// jumpers定义了每个字元处的向前跳转信息，包括这个跳转对应的分词，
	// 以及从文本段开始到该字元的最短路径值
	jumpers := make([]jumper, len(text))

	tokens := make([]*Token, seg.dict.maxTokenLength)

	for current := 0; current < len(text); current++ {
		// 找到前一个字元处的最短路径，以便计算后续路径值
		var baseDistance float32
		if current == 0 {
			// 当本字元在文本首部时，基础距离应该是零
			baseDistance = 0
		} else {
			baseDistance = jumpers[current-1].minDistance
		}

		// 寻找所有以当前字元开头的分词
		numTokens := seg.dict.lookupTokens(
			text[current:minInt(current+seg.dict.maxTokenLength, len(text))], tokens)

		// 对所有可能的分词，更新分词结束字元处的跳转信息
		for iToken := 0; iToken < numTokens; iToken++ {
			location := current + len(tokens[iToken].text) - 1
			if !searchMode || current != 0 || location != len(text)-1 {

				updateJumper(&jumpers[location], baseDistance, tokens[iToken])
			}

		}

		// 当前字元没有对应分词时补加一个伪分词
		if numTokens == 0 || len(tokens[0].text) > 1 {
			updateJumper(&jumpers[current], baseDistance,
				&Token{text: []Text{text[current]}, frequency: 1, distance: 32, pos: "x"})
		}

	}

	// 从后向前扫描第一遍得到需要添加的分词数目
	numSeg := 0
	for index := len(text) - 1; index >= 0; {
		location := index - len(jumpers[index].token.text) + 1
		numSeg++
		index = location - 1
	}

	// 从后向前扫描第二遍添加分词到最终结果
	outputSegments := make([]Segment, numSeg)

	for index := len(text) - 1; index >= 0; {
		location := index - len(jumpers[index].token.text) + 1
		numSeg--
		outputSegments[numSeg].token = jumpers[index].token
		index = location - 1

	}

	// 计算各个分词的字节位置
	bytePosition := 0
	for iSeg := 0; iSeg < len(outputSegments); iSeg++ {
		outputSegments[iSeg].start = bytePosition
		bytePosition += textSliceByteLength(outputSegments[iSeg].token.text)
		outputSegments[iSeg].end = bytePosition
	}
	return outputSegments
}

// 更新跳转信息:
// 	1. 当该位置从未被访问过时(jumper.minDistance为零的情况)，或者
//	2. 当该位置的当前最短路径大于新的最短路径时
// 将当前位置的最短路径值更新为baseDistance加上新分词的概率
func updateJumper(jumper *jumper, baseDistance float32, token *Token) {
	newDistance := baseDistance + token.distance
	if jumper.minDistance == 0 || jumper.minDistance > newDistance {
		jumper.minDistance = newDistance
		jumper.token = token
	}
}
