package commands

import (
	"bytes"
	"fmt"
	"github.com/anhgelus/gokord/cmd"
	"github.com/anhgelus/gokord/logger"
	"github.com/bwmarrin/discordgo"
	"github.com/nyttikord/nerdkord/latex"
)

const (
	LaTeXModalID = "latex_modal"
)

func OnLatexModalSubmit(s *discordgo.Session, i *discordgo.InteractionCreate, data discordgo.ModalSubmitInteractionData, resp *cmd.ResponseBuilder) {
	latexSource := data.Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
	latex.RenderLatexAndReply(s, i, resp, latexSource, latex.GetSourceID)
}

func OnSourceButton(_ *discordgo.Session, i *discordgo.InteractionCreate, _ discordgo.MessageComponentInteractionData, resp *cmd.ResponseBuilder) {
	resp.IsEphemeral()
	k := fmt.Sprintf("%s:%s", i.ChannelID, i.Message.ID)
	source, ok := latex.GetSource(k)
	if !ok {
		logger.Warn("cannot find source", "key", k)
		if err := resp.SetMessage("Cannot find the source").Send(); err != nil {
			logger.Alert("commands/latex.go - Sending error cannot find source", err.Error())
		}
		return
	}

	msg := fmt.Sprintf("Latex source:\n```\n%s\n```", *source)
	if len(msg) > 1999 {
		resp.SetMessage("Latex source:").AddFile(&discordgo.File{
			Name:        "source.tex",
			ContentType: "application/x-latex",
			Reader:      bytes.NewBuffer([]byte(*source)),
		})
	} else {
		resp.SetMessage(msg)
	}
	if err := resp.Send(); err != nil {
		logger.Alert("commands/latex.go - Sending source", err.Error())
	}
}

func Latex(s *discordgo.Session, i *discordgo.InteractionCreate, o cmd.OptionMap, resp *cmd.ResponseBuilder) {
	source, ok := o["source"]
	if ok {
		latex.RenderLatexAndReply(s, i, resp, source.StringValue(), latex.GetSourceID)
		return
	}
	err := resp.SetCustomID(LaTeXModalID).
		IsModal().
		SetTitle("LaTeX compiler").
		AddComponent(discordgo.ActionsRow{Components: []discordgo.MessageComponent{
			discordgo.TextInput{
				CustomID:    "source",
				Label:       "Source",
				Style:       discordgo.TextInputParagraph,
				Placeholder: "Did you know $1 + 1 = 2$ ?",
				Required:    true,
				MinLength:   0,
				MaxLength:   4000,
			},
		}}).Send()
	if err != nil {
		logger.Alert("commands/latex.go - Sending modal", err.Error())
	}
}
