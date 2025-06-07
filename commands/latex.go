package commands

import (
	"github.com/anhgelus/gokord/utils"
	"github.com/bwmarrin/discordgo"
	"github.com/nyttikord/nerdkord/libs/img"
	"github.com/nyttikord/nerdkord/libs/latex2png"
	"image/color"
	"image/png"
	"os"
)

func OnLatexModalSubmit(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionModalSubmit {
		return
	}

	resp := utils.NewResponseBuilder(s, i).IsDeferred().IsEphemeral()
	err := resp.Send()
	if err != nil {
		utils.SendAlert("commands/latex.go - Sending deferred", err.Error())
		return
	}

	data := i.ModalSubmitData()
	if data.CustomID != "latex_modal" {
		utils.SendDebug("commands/latex.go - Unknown modal ID")
		return
	}

	latexSource := data.Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value

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
		utils.SendDebug("commands/latex.go - Error while compiling latex")
		err = resp.SetMessage("Error while compiling latex").Send()
		if err != nil {
			utils.SendAlert("commands/latex.go - Sending latex error", err.Error())
		}
		return
	}

	latexImage, err := png.Decode(file)
	if err != nil {
		utils.SendAlert("commands/latex.go - Error while decoding dvipng output image", err.Error())
		err = resp.
			SetMessage("An error occurred while running this command. Try again later, or contact a bot developer").
			Send()
		if err != nil {
			utils.SendAlert("commands/latex.go - Sending decoding error", err.Error())
		}
		return
	}
	name := file.Name()
	_ = file.Close()

	output, err := os.Create(name)
	err = png.Encode(output, img.Pad(latexImage, 5))
	if err != nil {
		utils.SendAlert("commands/latex.go - Error while encoding padded image", err.Error())
		err = resp.
			SetMessage("An error occurred while running this command. Try again later, or contact a bot developer").
			Send()
		if err != nil {
			utils.SendAlert("commands/latex.go - Sending encoding error", err.Error())
		}
		return
	}

	err = resp.AddFile(&discordgo.File{
		Name:        "generated_latex.png",
		ContentType: "image/png",
		Reader:      output,
	}).IsEdit().Send()
	if err != nil {
		utils.SendAlert("commands/latex.go - Sending latex", err.Error())
	}

	_ = output.Close()
}

func Latex(s *discordgo.Session, i *discordgo.InteractionCreate, _ utils.OptionMap, _ *utils.ResponseBuilder) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			CustomID: "latex_modal",
			Title:    "Latex compiler",
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "source",
							Label:       "Source",
							Style:       discordgo.TextInputParagraph,
							Placeholder: "Did you know $1 + 1 = 2$ ?",
							Required:    true,
							MinLength:   0,
							MaxLength:   4000,
						},
					},
				},
			},
		},
	})
	if err != nil {
		utils.SendAlert("commands/latex.go - Sending modal", err.Error())
	}
}
