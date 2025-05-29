package commands

import (
	"github.com/anhgelus/gokord/utils"
	"github.com/bwmarrin/discordgo"
	"github.com/nyttikord/nerdkord/libs/latex2png"
	"image/color"
)

func Latex(s *discordgo.Session, i *discordgo.InteractionCreate) {
	resp := utils.NewResponseBuilder(s, i).IsDeferred().IsEphemeral()
	err := resp.Send()
	if err != nil {
		utils.SendAlert("commands/latex.go - Sending deferred", err.Error())
		return
	}

	latexSource := i.Interaction.ApplicationCommandData().Options[0].StringValue()

	file, err := latex2png.Compile(latexSource, &latex2png.Options{
		LatexBinary:      "latex",
		DvipngBinary:     "dvipng",
		PreambleFilePath: "config/defaultPreamble.tex",
		AddBeginDocument: true,
		OutputFormat:     latex2png.PNG,
		BackgroundColor:  color.RGBA{R: 54, G: 57, B: 62, A: 255},
		ForegroundColor:  color.White,
		ImageDPI:         300,
	})

	if err != nil {
		resp.Message("Error while compiling latex")
		err = resp.Send()
		if err != nil {
			utils.SendAlert("commands/latex.go - Sending error", err.Error())
		}
		return
	}

	_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Files: []*discordgo.File{{
			Name:        "generated_latex.png",
			ContentType: "image/png",
			Reader:      file,
		}},
	})
	if err != nil {
		utils.SendAlert("commands/latex.go - Sending latex", err.Error())
	}
}
