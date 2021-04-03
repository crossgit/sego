package sego

func (seg *Segmenter) cutAll(text []Text) []CutAll {
	// log.Println("~~~~~~~~", textSliceToString(text))
	// if len(text) == 1 {
	// 	return []CutAll{}
	// }

	// jumpers定义了每个字元处的向前跳转信息，包括这个跳转对应的分词，
	// 以及从文本段开始到该字元的最短路径值
	jumpers := make([]jumper, len(text))
	tokens := make([]*Token, seg.dict.maxTokenLength)
	cutAllArray := make([]CutAll, len(text))

	start := 0
	for k, v := range text {
		cutAllArray[k].Start = start
		cutAllArray[k].End = start + len([]rune(string(v)))
		cutAllArray[k].Token = string(v)
		cutAllArray[k].Pos = "x"
		start += len([]rune(string(v)))
	}

	// log.Println("-----", textSliceToString(text))
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

			updateJumper1(&jumpers[location], baseDistance, tokens[iToken])

			cutAllArray[current].End = cutAllArray[current].Start + tokens[iToken].Length()
			cutAllArray[current].Token = tokens[iToken].Text()
			cutAllArray[current].Pos = tokens[iToken].Pos()

		}

	}

	result := make([]CutAll, 0)
	for _, v := range cutAllArray {
		if v.Pos != "x" {
			result = append(result, v)
		}
	}

	return result
}

// updateJumper1 全切的跳转,就跳1位
func updateJumper1(jumper *jumper, baseDistance float32, token *Token) {
	newDistance := baseDistance + token.distance
	if jumper.minDistance == 0 {
		jumper.minDistance = newDistance
		jumper.token = token
	}

}
