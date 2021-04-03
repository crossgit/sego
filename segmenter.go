//Go中文分词
package sego

import (
	"log"
	"unicode"
	"unicode/utf8"
)

const (
	minTokenFrequency = 2 // 仅从字典文件中读取大于等于此频率的分词
)

// 分词器结构体
type Segmenter struct {
	dict *Dictionary
}

// 该结构体用于记录Viterbi算法中某字元处的向前分词跳转信息
type jumper struct {
	minDistance float32
	token       *Token
}

// 返回分词器使用的词典
func (seg *Segmenter) Dictionary() *Dictionary {
	return seg.dict
}

// 对文本分词
//
// 输入参数：
//	bytes	UTF8文本的字节数组
//
// 输出：
//	[]Segment	划分的分词
func (seg *Segmenter) Segment(bytes []byte) []Segment {
	// return seg.internalSegment(bytes, false)
	// 处理特殊情况
	if len(bytes) == 0 {
		return []Segment{}
	}

	// 划分字元
	text := splitTextToWords(bytes)
	log.Println("internalSegment:", textSliceToString(text))
	return seg.cutJump(text, false)
}

// CutAll 逐字全切结构
type CutAll struct {
	Start int
	End   int
	Token string
	Pos   string
}

// SegmentAll 全切情况
func (seg *Segmenter) SegmentAll(bytes []byte) []CutAll {
	// 处理特殊情况
	if len(bytes) == 0 {
		return []CutAll{}
	}

	// 划分字元
	text := splitTextToWords(bytes)
	return seg.cutAll(text)
}

func (seg *Segmenter) InternalSegment(bytes []byte, searchMode bool) []Segment {
	return seg.internalSegment(bytes, searchMode)
}

func (seg *Segmenter) internalSegment(bytes []byte, searchMode bool) []Segment {
	// 处理特殊情况
	if len(bytes) == 0 {
		return []Segment{}
	}

	// 划分字元
	text := splitTextToWords(bytes)
	log.Println("internalSegment:")
	return seg.cutJump(text, searchMode)
}

// 取两整数较小值
func minInt(a, b int) int {
	if a > b {
		return b
	}
	return a
}

// 取两整数较大值
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// 将文本划分成字元
func splitTextToWords(text Text) []Text {
	output := make([]Text, 0, len(text)/3)
	current := 0
	inAlphanumeric := true
	alphanumericStart := 0
	for current < len(text) {
		r, size := utf8.DecodeRune(text[current:])
		if size <= 2 && (unicode.IsLetter(r) || unicode.IsNumber(r)) {
			// 当前是拉丁字母或数字（非中日韩文字）
			if !inAlphanumeric {
				alphanumericStart = current
				inAlphanumeric = true
			}
		} else {
			if inAlphanumeric {
				inAlphanumeric = false
				if current != 0 {
					output = append(output, toLower(text[alphanumericStart:current]))
				}
			}
			output = append(output, text[current:current+size])
		}
		current += size
	}

	// 处理最后一个字元是英文的情况
	if inAlphanumeric {
		if current != 0 {
			output = append(output, toLower(text[alphanumericStart:current]))
		}
	}

	return output
}

// 将英文词转化为小写
func toLower(text []byte) []byte {
	output := make([]byte, len(text))
	for i, t := range text {
		if t >= 'A' && t <= 'Z' {
			output[i] = t - 'A' + 'a'
		} else {
			output[i] = t
		}
	}
	return output
}
