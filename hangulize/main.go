package main

import (
	"github.com/hangulize/hangulize"
	"github.com/hangulize/hangulize/phonemize/furigana"
	"github.com/hangulize/hangulize/phonemize/pinyin"

	"github.com/hangulize/cmd/hangulize/cmd"
)

func init() {
	hangulize.UsePhonemizer(&furigana.P)
	hangulize.UsePhonemizer(&pinyin.P)
}

func main() {
	cmd.Execute()
}
