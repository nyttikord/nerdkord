package latex

import (
	"github.com/bwmarrin/discordgo"
	"regexp"
)

var (
	regexDetectLatexDollar    = regexp.MustCompile(`\$[^ ]+.*\$`)
	regexDetectLatexOneLine   = regexp.MustCompile(`\\\([^ ]+.*\\\)`)
	regexDetectLatexMultiLine = regexp.MustCompile(`\\\[[^ ]+.*\\]`)
	regexDetectLatexBegEnd    = regexp.MustCompile(`\\begin[\n ]*\{.+}(\n|.)*\\end[\n ]*\{.+}`)
)

func HandleLatexSourceCode(s *discordgo.Session, m *discordgo.MessageCreate) {
	msg := m.Content
	if !regexDetectLatexDollar.MatchString(msg) &&
		!regexDetectLatexOneLine.MatchString(msg) &&
		!regexDetectLatexMultiLine.MatchString(msg) &&
		!regexDetectLatexBegEnd.MatchString(msg) {
		return
	}
	//TODO: render latex
}
