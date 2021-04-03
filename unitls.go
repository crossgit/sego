package sego

import (
	"bytes"
	"fmt"
	"log"
	"strings"
)

// 输出分词结果为字符串
//
// 有两种输出模式，以"中华人民共和国"为例
//
//  普通模式（searchMode=false）输出一个分词"中华人民共和国/ns "
//  搜索模式（searchMode=true） 输出普通模式的再细致切分：
//      "中华/nz 人民/n 共和/nz 共和国/ns 人民共和国/nt 中华人民共和国/ns "
//
// 搜索模式主要用于给搜索引擎提供尽可能多的关键字，详情请见Token结构体的注释。
func SegmentsToString(segs []Segment, searchMode bool) (output string) {
	if searchMode {
		for _, seg := range segs {
			output += tokenToString(seg.token)
		}
	} else {
		for _, seg := range segs {
			output += fmt.Sprintf(
				"%s/%s ", textSliceToString(seg.token.text), seg.token.pos)
		}
	}
	return
}

func tokenToString(token *Token) (output string) {
	hasOnlyTerminalToken := true
	for _, s := range token.segments {
		if len(s.token.segments) > 1 {
			hasOnlyTerminalToken = false
		}
	}

	if !hasOnlyTerminalToken {
		for _, s := range token.segments {
			if s != nil {
				output += tokenToString(s.token)
			}
		}
	}
	output += fmt.Sprintf("%s/%s ", textSliceToString(token.text), token.pos)
	return
}

// 输出分词结果到一个字符串slice
//
// 有两种输出模式，以"中华人民共和国"为例
//
//  普通模式（searchMode=false）输出一个分词"[中华人民共和国]"
//  搜索模式（searchMode=true） 输出普通模式的再细致切分：
//      "[中华 人民 共和 共和国 人民共和国 中华人民共和国]"
//
// 搜索模式主要用于给搜索引擎提供尽可能多的关键字，详情请见Token结构体的注释。

func SegmentsToSlice(segs []Segment, searchMode bool) (output []string) {
	if searchMode {
		for _, seg := range segs {
			output = append(output, tokenToSlice(seg.token)...)
		}
	} else {
		for _, seg := range segs {
			output = append(output, seg.token.Text())
		}
	}
	return
}

func tokenToSlice(token *Token) (output []string) {
	hasOnlyTerminalToken := true
	for _, s := range token.segments {
		if len(s.token.segments) > 1 {
			hasOnlyTerminalToken = false
		}
	}
	if !hasOnlyTerminalToken {
		for _, s := range token.segments {
			output = append(output, tokenToSlice(s.token)...)
		}
	}
	output = append(output, textSliceToString(token.text))
	return output
}

// 将多个字元拼接一个字符串输出
func textSliceToString(text []Text) string {
	return Join(text)
}

func Join(a []Text) string {
	switch len(a) {
	case 0:
		return ""
	case 1:
		return string(a[0])
	case 2:
		// Special case for common small values.
		// Remove if golang.org/issue/6714 is fixed
		return string(a[0]) + string(a[1])
	case 3:
		// Special case for common small values.
		// Remove if golang.org/issue/6714 is fixed
		return string(a[0]) + string(a[1]) + string(a[2])
	}
	n := 0
	for i := 0; i < len(a); i++ {
		n += len(a[i])
	}

	b := make([]byte, n)
	bp := copy(b, a[0])
	for _, s := range a[1:] {
		bp += copy(b[bp:], s)
	}
	return string(b)
}

// 返回多个字元的字节总长度
func textSliceByteLength(text []Text) (length int) {
	for _, word := range text {
		length += len([]rune(string(word)))
	}
	return
}

func textSliceToBytes(text []Text) []byte {
	var buf bytes.Buffer
	for _, word := range text {
		buf.Write(word)
	}
	return buf.Bytes()
}

// ------追加一种特殊情况
// 当添加的pos是u和p时,需要构造: 将最后的u和最开始的p放在一起,中间间隔offset个统配字.

// Output 输出格式,可以用ntoken与ptoken来确定一些其他属性
type Output struct {
	Start  int
	End    int
	ULeft  string
	UToken string
	XToken string
	PToken string
	PRight string
	Pos    string
}

// SegmentsOutput 输出到map里
func SegmentsOutput(segs []Segment, lOffset, mOffset, rOffset int) []Output {
	output := make([]Output, 0)
	segsLen := len(segs)
	for i := 0; i < segsLen-1; {
		nseg := segs[i]
		npos := nseg.token.pos
		xtoken := make([]string, 0)
		if npos == "u" {
			for j := i + 1; j < segsLen; j++ {
				pseg := segs[j]
				ppos := pseg.token.pos
				if ppos == "u" {
					i = j
					break
				} else if ppos == "p" {
					nend := nseg.End()
					pstart := pseg.Start()
					if pstart-nend > mOffset {
						i = j + 1
						break
					}
					var info Output
					info.Start = nseg.Start()
					info.End = pseg.End()
					info.Pos = ""
					info.UToken = tokenToStr(nseg.token)
					info.XToken = strings.Join(xtoken, "")
					info.PToken = tokenToStr(pseg.token)
					li := j - len(xtoken) - 1
					// log.Println(info.Start, info.End, li, j)
					if li > lOffset {
						info.ULeft = segmentArrayToStr(segs[li-lOffset : li])
					} else {
						info.ULeft = segmentArrayToStr(segs[:li])
					}
					if j+rOffset < segsLen {
						info.PRight = segmentArrayToStr(segs[j+1 : j+1+rOffset])
					} else {
						info.PRight = segmentArrayToStr(segs[j+1:])
					}
					output = append(output, info)
					i = j + 1
					break
				} else {
					xseg := segs[j]
					if len(xtoken) < mOffset {
						xtoken = append(xtoken, tokenToStr(xseg.token))
					}
					i++
				}
			}
		} else {
			i++
		}
	}
	return output
}

type OutputSingle struct {
	Start  int
	End    int
	NLeft  string
	NToken string
	NRight string
	Pos    string
}

// SegmentsOutputSingle 单一的输出
// 输入 左右边距
func SegmentsOutputSingle(segs []Segment, lOffset, rOffset int) []OutputSingle {
	output := make([]OutputSingle, 0)
	segsLen := len(segs)
	for i := 0; i < segsLen; i++ {
		nseg := segs[i]
		npos := nseg.token.pos
		if npos == "n" {
			var info OutputSingle
			info.Start = nseg.Start()
			info.End = nseg.End()
			info.Pos = npos
			info.NToken = tokenToStr(nseg.token)
			if i > lOffset {
				info.NLeft = segmentArrayToStr(segs[i-lOffset : i])
			} else {
				info.NLeft = segmentArrayToStr(segs[:i])
			}
			if i+rOffset < segsLen {
				info.NRight = segmentArrayToStr(segs[i+1 : i+1+rOffset])
			} else {
				info.NRight = segmentArrayToStr(segs[i+1:])
			}
			output = append(output, info)

		}
	}
	return output
}

// SegmentsOutputAll 单一的输出
// 输入 左右边距
func SegmentsOutputAll(segs []Segment, lOffset, rOffset int) []OutputSingle {
	output := make([]OutputSingle, 0)
	segsLen := len(segs)
	for i := 0; i < segsLen; i++ {
		nseg := segs[i]
		log.Println("获取的时候:", tokenToStr(nseg.token))
		npos := nseg.token.pos
		if npos != "x" {
			var info OutputSingle
			info.Start = nseg.Start()
			info.End = nseg.End()
			info.Pos = npos
			info.NToken = tokenToStr(nseg.token)
			if i > lOffset {
				info.NLeft = segmentArrayToStr(segs[i-lOffset : i])
			} else {
				info.NLeft = segmentArrayToStr(segs[:i])
			}
			if i+rOffset < segsLen {
				info.NRight = segmentArrayToStr(segs[i+1 : i+1+rOffset])
			} else {
				info.NRight = segmentArrayToStr(segs[i+1:])
			}
			output = append(output, info)

		}
	}
	return output
}

// SegmentsOutputSingleAll 单一的输出
// 输入 忽略左右边距,去除x,全部输出
func SegmentsOutputSingleAll(segs []Segment) []OutputSingle {
	output := make([]OutputSingle, 0)
	segsLen := len(segs)
	for i := 0; i < segsLen; i++ {
		nseg := segs[i]
		npos := nseg.token.pos
		if npos != "x" && npos[:1] != "d" {
			var info OutputSingle
			info.Start = nseg.Start()
			info.End = nseg.End()
			info.Pos = npos
			info.NToken = tokenToStr(nseg.token)
			output = append(output, info)
		}
	}
	return output
}

// SegmentsJumpOutput 单一的输出
// 输入 忽略左右边距,去除x,全部输出
func SegmentsJumpOutput(segs []CutAll) []OutputSingle {
	output := make([]OutputSingle, 0)
	segsLen := len(segs)
	if segsLen == 0 {
		return output
	}
	index := 0
	for i := 0; i < segsLen; i++ {
		nseg := segs[i]
		if i == 0 {
			output = append(output, OutputSingle{
				Start:  nseg.Start,
				End:    nseg.End,
				NToken: nseg.Token,
				Pos:    nseg.Pos,
			})
			index = nseg.End
			continue
		}
		if nseg.Start < index {
			continue
		}

		output = append(output, OutputSingle{
			Start:  nseg.Start,
			End:    nseg.End,
			NToken: nseg.Token,
			Pos:    nseg.Pos,
		})
		index = nseg.End

	}
	return output
}

// SegmentsJumpOutput 单一的输出
// 输入 忽略左右边距,去除x,全部输出
func SegmentsAllOutput(segs []CutAll) []OutputSingle {
	output := make([]OutputSingle, 0)
	segsLen := len(segs)
	if segsLen == 0 {
		return output
	}
	for i := 0; i < segsLen; i++ {
		nseg := segs[i]

		output = append(output, OutputSingle{
			Start:  nseg.Start,
			End:    nseg.End,
			NToken: nseg.Token,
			Pos:    nseg.Pos,
		})
	}
	return output
}

// SegmentsMergeOutput 全切的合并后输出
// 输入 忽略左右边距,去除x,全部输出
func SegmentsMergeOutput(segs []CutAll) []OutputSingle {
	// log.Println("segs0:", segs)
	output := make([]OutputSingle, 0)
	segsLen := len(segs)
	if segsLen == 0 {
		return output
	}

	indexSeg := segs[0]
	output = append(output, OutputSingle{
		Start:  indexSeg.Start,
		End:    indexSeg.End,
		NToken: indexSeg.Token,
		Pos:    indexSeg.Pos,
	})
	if segsLen == 1 {
		return output
	}
	for i := 0; i < segsLen; i++ {
		lastOne := output[len(output)-1] //合并最后一个
		nseg := segs[i]

		if nseg.Start > lastOne.End {
			if lastOne.Pos[:1] == "d" {
				output[len(output)-1].Start = nseg.Start
				output[len(output)-1].End = nseg.End
				output[len(output)-1].NToken = nseg.Token
				output[len(output)-1].Pos = nseg.Pos
			} else {
				output = append(output, OutputSingle{
					Start:  nseg.Start,
					End:    nseg.End,
					NToken: nseg.Token,
					Pos:    nseg.Pos,
				})
			}
		} else {
			if lastOne.End > nseg.Start && nseg.End > lastOne.End {
				// 相交
				if nseg.Pos[:1] == "d" {
					// 要删除的词都不添加,并且前一个词也会变成要删除词
					output[len(output)-1].Pos = nseg.Pos
					continue
				}
				if lastOne.Pos[:1] != "d" && nseg.Pos == lastOne.Pos {
					// 向后延续 // ABCD:n CDEF:n -> ABCDEF:n
					output[len(output)-1].End = nseg.End
					output[len(output)-1].NToken = lastOne.NToken + string([]rune(nseg.Token)[lastOne.End-nseg.Start:])
					continue
				}
				output[len(output)-1].Start = nseg.Start
				output[len(output)-1].End = nseg.End
				output[len(output)-1].NToken = nseg.Token
				output[len(output)-1].Pos = nseg.Pos

				// 删除词和一般词拼接,还是删除词
				continue
			}
			if lastOne.End == nseg.Start && nseg.End > lastOne.End {
				// 不相交
				if nseg.Pos[:1] != "d" {
					if nseg.Pos == lastOne.Pos {
						// 向后延续 // ABCD:n EFGH:n -> ABCDEFGH:n
						output[len(output)-1].End = nseg.End
						output[len(output)-1].NToken = lastOne.NToken + string([]rune(nseg.Token)[lastOne.End-nseg.Start:])
						continue
					}
					output = append(output, OutputSingle{
						Start:  nseg.Start,
						End:    nseg.End,
						NToken: nseg.Token,
						Pos:    nseg.Pos,
					})

					continue
				}
				continue
			}

			continue
		}
	}
	// log.Println("==", output)
	return output
}

func segmentArrayToStr(segs []Segment) (output string) {
	for _, v := range segs {
		output += tokenToStr(v.token)
	}
	return output
}

// SegmentsOutputArray 只输出存在的词
func SegmentsOutputArray(segs []Segment) []string {
	output := make([]string, 0)
	segsLen := len(segs)
	for i := 0; i < segsLen; i++ {
		nseg := segs[i]
		npos := nseg.token.pos
		if npos == "n" {
			output = append(output, tokenToStr(nseg.token))

		}
	}
	return output
}

func tokenToStr(token *Token) (output string) {
	// hasOnlyTerminalToken := true
	// for _, s := range token.segments {
	// 	if len(s.token.segments) > 1 {
	// 		hasOnlyTerminalToken = false
	// 	}
	// }

	// if !hasOnlyTerminalToken {
	// 	for _, s := range token.segments {
	// 		if s != nil {
	// 			output += tokenToStr(s.token)
	// 		}
	// 	}
	// }
	output += textSliceToString(token.text)
	return
}
