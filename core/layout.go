package core

import "strings"

type AlignH int
type AlignV int

const (
	AlignLeft AlignH = iota
	AlignCenter
	AlignRight
)

const (
	AlignTop AlignV = iota
	AlignMiddle
	AlignBottom
)

func Center(block string, boxW, boxH int) string {
	return PlaceBlock(block, boxW, boxH, AlignCenter, AlignMiddle)
}


func PlaceBlock(block string, boxW, boxH int, h AlignH, v AlignV) string {
	if boxW <= 0 || boxH <= 0 || block == "" {
		return block
	}
	lines := strings.Split(block, "\n")
	_, bh := blockSize(lines)

	topPad := 0
	switch v {
	case AlignTop:
		topPad = 0
	case AlignMiddle:
		if boxH > bh {
			topPad = (boxH - bh) / 2
		}
	case AlignBottom:
		if boxH > bh {
			topPad = boxH - bh
		}
	}

	var b strings.Builder
	if topPad > 0 {
		b.WriteString(strings.Repeat("\n", topPad))
	}

	for i, line := range lines {
		leftPad := 0
		lw := displayWidth(line)
		switch h {
		case AlignLeft:
			leftPad = 0
		case AlignCenter:
			if boxW > lw {
				leftPad = (boxW - lw) / 2
			}
		case AlignRight:
			if boxW > lw {
				leftPad = boxW - lw
			}
		}
		if leftPad > 0 {
			b.WriteString(strings.Repeat(" ", leftPad))
		}
		b.WriteString(line)
		if i < len(lines)-1 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}

func blockSize(lines []string) (w, h int) {
	h = len(lines)
	for _, ln := range lines {
		if dw := displayWidth(ln); dw > w {
			w = dw
		}
	}
	return
}


func displayWidth(s string) int {
	plain := StripANSI(s)
	w := 0
	for _, r := range plain {
		if r == '\t' {
			next := 4 - (w % 4)
			w += next
			continue
		}
		w++
	}
	return w
}
