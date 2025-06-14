package commands

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/anhgelus/gokord/utils"
	"github.com/bwmarrin/discordgo"
	"github.com/nyttikord/nerdkord/data"
	"github.com/nyttikord/nerdkord/libs/img"
	"github.com/nyttikord/nerdkord/libs/latex2png"
	"image/color"
	"image/png"
	"math"
	"time"
)

var (
	bgColor = color.RGBA{R: 54, G: 57, B: 62, A: 255}
	fgColor = color.White

	defaultPreprocessingOptions = &latex2png.PreprocessingOptions{
		ForbiddenCommands:           []string{"include", "import"},
		CommandsBeforeBeginDocument: []string{"usepackage"},
		TemplateFile:                "config/template.tex",
	}

	sourceMap = make(map[string]*string, 100)
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
	err := resp.Send()
	if err != nil {
		utils.SendAlert("commands/latex.go - Sending deferred", err.Error())
		return
	}
	resp.IsEphemeral()

	var u *discordgo.User
	if i.User == nil {
		u = i.Member.User
	} else {
		u = i.User
	}

	nerd, err := data.GetNerd(u.ID)
	if err != nil {
		utils.SendAlert("commands/latex.go - Getting nerd", err.Error(), "discord_id", u.ID)
		if err = resp.SetMessage("Error while getting your profile. Please report the bug.").Send(); err != nil {
			utils.SendAlert("commands/latex.go - Sending error getting nerd", err.Error())
		}
		return
	}

	latexSource := submitData.Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value

	file := new(bytes.Buffer)
	opt := &*defaultPreprocessingOptions
	opt.UserPreamble = nerd.Preamble
	err = latex2png.Compile(file, latexSource, &latex2png.Options{
		LatexBinary:          "latex",
		DvipngBinary:         "dvipng",
		OutputFormat:         latex2png.PNG,
		BackgroundColor:      bgColor,
		ForegroundColor:      fgColor,
		ImageDPI:             300,
		PreprocessingOptions: opt,
	})

	if err != nil {
		if errors.Is(err, latex2png.ErrPreprocessor) {
			utils.SendDebug("commands.latex.go - Preprocessing error")
			err = resp.SetMessage("```\n" + err.Error() + "\n```").Send()
			if err != nil {
				utils.SendAlert("commands/latex.go - Sending preprocessing error", err.Error())
			}
			return
		}

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

	output := new(bytes.Buffer)
	err = png.Encode(output, img.Pad(latexImage, 5+int(math.Ceil(float64(latexImage.Bounds().Dx())*(1./100.))), bgColor))
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

	err = resp.NotEphemeral().AddFile(&discordgo.File{
		Name:        "generated_latex.png",
		ContentType: "image/png",
		Reader:      output,
	}).AddComponent(discordgo.ActionsRow{Components: []discordgo.MessageComponent{
		discordgo.Button{
			Label:    "Source",
			Style:    discordgo.SecondaryButton,
			Disabled: false,
			Emoji:    &discordgo.ComponentEmoji{Name: "📝"},
			CustomID: GetSourceID,
		},
	}}).Send()
	if err != nil {
		utils.SendAlert("commands/latex.go - Sending latex", err.Error())
		return
	}
	k := fmt.Sprintf("%s:%d", i.ChannelID, time.Now().Unix())
	sourceMap[k] = &latexSource
	// remove source button after 15 minutes and clean map
	go func(resp *utils.ResponseBuilder, k string, output *bytes.Buffer) {
		time.Sleep(15 * time.Minute)
		err := resp.IsEdit().AddFile(&discordgo.File{
			Name:        "generated_latex.png",
			ContentType: "image/png",
			Reader:      output,
		}).Send()
		if err != nil {
			utils.SendAlert("commands/latex.go - Cannot remove source button", err.Error())
		}
		delete(sourceMap, k)
	}(resp, k, output)
}

func OnSourceButton(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionMessageComponent {
		return
	}

	submitData := i.ModalSubmitData()
	if submitData.CustomID != GetSourceID {
		utils.SendDebug("commands/latex.go - not a source button ID")
		return
	}
	resp := utils.NewResponseBuilder(s, i).IsDeferred()
	t, err := discordgo.SnowflakeTimestamp(i.Message.ID)
	if err != nil {
		utils.SendAlert("commands/latex.go - Getting time from discord", err.Error(), "id", i.Message.ID)
		if err = resp.SetMessage("Error while getting the source").Send(); err != nil {
			utils.SendAlert("commands/latex.go - Sending error getting time from discord", err.Error())
		}
		return
	}
	k := fmt.Sprintf("%s:%d", i.ChannelID, t.Unix())
	source, ok := sourceMap[k]
	if !ok {
		utils.SendWarn("cannot find source", "key", k)
		if err = resp.SetMessage("Cannot find the source").Send(); err != nil {
			utils.SendAlert("commands/latex.go - Sending error cannot find source", err.Error())
		}
		return
	}
	if err = resp.SetMessage(fmt.Sprintf("```\n%s\n```", *source)).Send(); err != nil {
		utils.SendAlert("commands/latex.go - Sending source", err.Error())
	}
}

func Latex(_ *discordgo.Session, _ *discordgo.InteractionCreate, _ utils.OptionMap, resp *utils.ResponseBuilder) {
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
