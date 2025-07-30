package commands

import (
	"bytes"
	"fmt"
	"github.com/anhgelus/gokord/utils"
	"github.com/bwmarrin/discordgo"
	"github.com/nyttikord/nerdkord/latex"
)

const (
	LaTeXModalID = "latex_modal"
	GetSourceID  = "latex_source"
)

func OnLatexModalSubmit(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionModalSubmit {
		return
	}

	submitData := i.ModalSubmitData()
	if submitData.CustomID != LaTeXModalID {
		utils.SendDebug("commands/latex.go - not a latex modal ID")
		return
	}

	resp := utils.NewResponseBuilder(s, i).IsDeferred()

	latexSource := submitData.Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value

	latex.RenderLatex(s, i, resp, latexSource, GetSourceID)
}

func OnSourceButton(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionMessageComponent {
		return
	}

	submitData := i.MessageComponentData()
	if submitData.CustomID != GetSourceID {
		utils.SendDebug("commands/latex.go - not a source button ID")
		return
	}
	resp := utils.NewResponseBuilder(s, i).IsEphemeral()
	k := fmt.Sprintf("%s:%s", i.ChannelID, i.Message.ID)
	source, ok := latex.GetSource(k)
	if !ok {
		utils.SendWarn("cannot find source", "key", k)
		if err := resp.SetMessage("Cannot find the source").Send(); err != nil {
			utils.SendAlert("commands/latex.go - Sending error cannot find source", err.Error())
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
		utils.SendAlert("commands/latex.go - Sending source", err.Error())
	}
}

func Latex(s *discordgo.Session, i *discordgo.InteractionCreate, o utils.OptionMap, resp *utils.ResponseBuilder) {
	source, ok := o["source"]
	if ok {
		latex.RenderLatex(s, i, resp, source.StringValue(), GetSourceID)
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
		utils.SendAlert("commands/latex.go - Sending modal", err.Error())
	}
}
