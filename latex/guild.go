package latex

import (
	"fmt"
	"github.com/anhgelus/gokord/utils"
	"github.com/bwmarrin/discordgo"
	"regexp"
)

var (
	regexDetectLatexDollar    = regexp.MustCompile(`\$[^ ]+.*\$`)
	regexDetectLatexOneLine   = regexp.MustCompile(`\\\([^ ]+.*\\\)`)
	regexDetectLatexMultiLine = regexp.MustCompile(`\\\[[^ ]+.*\\]`)
	regexDetectLatexBegEnd    = regexp.MustCompile(`\\begin[\n ]*\{.+}(\n|.)*\\end[\n ]*\{.+}`)
)

func HandleLatexSourceCode(s *discordgo.Session, m *discordgo.MessageCreate, getSourceID string) {
	source := m.Content
	if !regexDetectLatexDollar.MatchString(source) &&
		!regexDetectLatexOneLine.MatchString(source) &&
		!regexDetectLatexMultiLine.MatchString(source) &&
		!regexDetectLatexBegEnd.MatchString(source) {
		return
	}
	output, err := RenderLatex(m.Author, source)
	if err != nil {
		//
		return
	}
	st, err := s.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{
		Content: fmt.Sprintf("**%s**", m.Author.DisplayName()),
		Components: []discordgo.MessageComponent{
			discordgo.Button{
				Label:    "",
				Style:    discordgo.SecondaryButton,
				Disabled: false,
				Emoji:    &discordgo.ComponentEmoji{Name: "📝"},
				CustomID: getSourceID,
			},
		},
		Files: []*discordgo.File{{
			Name:        "generated_latex.png",
			ContentType: "image/png",
			Reader:      output,
		}},
	})
	if err != nil {
		utils.SendAlert("latex/guild.go - Sending latex", err.Error())
		return
	}
	saveSourceWithMessage(s, st, source)
}
