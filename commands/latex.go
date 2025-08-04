package commands

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/anhgelus/gokord/cmd"
	"github.com/anhgelus/gokord/logger"
	"github.com/bwmarrin/discordgo"
	"github.com/nyttikord/nerdkord/db"
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
		logger.Debug("commands/latex.go - not a latex modal ID")
		return
	}

	resp := cmd.NewResponseBuilder(s, i).IsDeferred()

	latexSource := submitData.Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value

	renderLatex(s, i, resp, latexSource)
}

func renderLatex(s *discordgo.Session, i *discordgo.InteractionCreate, resp *cmd.ResponseBuilder, source string) {
	err := resp.Send()
	if err != nil {
		logger.Alert("commands/latex.go - Sending deferred", err.Error())
		return
	}
	resp.IsEphemeral()

	var u *discordgo.User
	if i.User == nil {
		u = i.Member.User
	} else {
		u = i.User
	}

	nerd, err := db.GetNerd(u.ID)
	if err != nil {
		logger.Alert("commands/latex.go - Getting nerd", err.Error(), "discord_id", u.ID)
		if err = resp.SetMessage("Error while getting your profile. Please report the bug.").Send(); err != nil {
			logger.Alert("commands/latex.go - Sending error getting nerd", err.Error())
		}
		return
	}

	file := new(bytes.Buffer)
	opt := &*defaultPreprocessingOptions
	opt.UserPreamble = nerd.Preamble
	err = latex2png.Compile(file, source, &latex2png.Options{
		LatexBinary:          "latex",
		DvipngBinary:         "dvipng",
		OutputFormat:         latex2png.PNG,
		BackgroundColor:      bgColor,
		ForegroundColor:      fgColor,
		ImageDPI:             300,
		PreprocessingOptions: opt,
	})

	if err != nil {
		if errors.As(err, &latex2png.ErrLatexCompilation{}) {
			logger.Debug("commands/latex.go - Latex compilation error")

			if len(err.Error()) > 1950 {
				resp.SetMessage("⚠️ Compilation error").AddFile(
					&discordgo.File{
						Name:        "error.txt",
						ContentType: "text/plain",
						Reader:      bytes.NewReader([]byte(err.Error())),
					},
				)
			} else {
				resp.SetMessage("⚠️ Compilation error:\n```\n" + err.Error() + "\n```")
			}
			err = resp.IsEphemeral().AddComponent(discordgo.ActionsRow{Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "",
					Style:    discordgo.SecondaryButton,
					Disabled: false,
					Emoji:    &discordgo.ComponentEmoji{Name: "📝"},
					CustomID: GetSourceID,
				},
			}}).Send()
			if err != nil {
				logger.Alert("commands/latex.go - Sending compilation error", err.Error())
			}

			saveSource(s, i, source)
			return
		}
		if errors.Is(err, latex2png.ErrPreprocessor) {
			logger.Debug("commands/latex.go - Preprocessing error")
			err = resp.SetMessage("```\n" + err.Error() + "\n```").Send()
			if err != nil {
				logger.Alert("commands/latex.go - Sending preprocessing error", err.Error())
			}
			return
		}

		logger.Alert("commands/latex.go - Compiling latex", err.Error())
		err = resp.SetMessage("Unexpected error while compiling latex").Send()
		if err != nil {
			logger.Alert("commands/latex.go - Sending unexpected latex error", err.Error())
		}
		return
	}

	latexImage, err := png.Decode(file)
	if err != nil {
		logger.Alert("commands/latex.go - Error while decoding dvipng output image", err.Error())
		err = resp.
			SetMessage("An error occurred while running this command. Try again later, or contact a bot developer").
			Send()
		if err != nil {
			logger.Alert("commands/latex.go - Sending decoding error", err.Error())
		}
		return
	}

	output := new(bytes.Buffer)
	err = png.Encode(output, img.Pad(latexImage, 5+int(math.Ceil(float64(latexImage.Bounds().Dx())*(1./100.))), bgColor))
	if err != nil {
		logger.Alert("commands/latex.go - Error while encoding padded image", err.Error())
		err = resp.
			SetMessage("An error occurred while running this command. Try again later, or contact a bot developer").
			Send()
		if err != nil {
			logger.Alert("commands/latex.go - Sending encoding error", err.Error())
		}
		return
	}

	err = resp.NotEphemeral().AddFile(&discordgo.File{
		Name:        "generated_latex.png",
		ContentType: "image/png",
		Reader:      output,
	}).AddComponent(discordgo.ActionsRow{Components: []discordgo.MessageComponent{
		discordgo.Button{
			Label:    "",
			Style:    discordgo.SecondaryButton,
			Disabled: false,
			Emoji:    &discordgo.ComponentEmoji{Name: "📝"},
			CustomID: GetSourceID,
		},
	}}).Send()
	if err != nil {
		logger.Alert("commands/latex.go - Sending latex", err.Error())
		return
	}
	// saving source
	saveSource(s, i, source)
}

func saveSource(s *discordgo.Session, i *discordgo.InteractionCreate, latexSource string) {
	m, err := s.InteractionResponse(i.Interaction)
	if err != nil {
		logger.Alert("commands/latex.go - Getting interaction response", err.Error(), "id", i.ID)
		return
	}
	k := fmt.Sprintf("%s:%s", i.ChannelID, m.ID)
	sourceMap[k] = &latexSource
	logger.Debug("source saved", "key", k)
	// remove source button after 5 minutes and clean map
	go func(s *discordgo.Session, i *discordgo.InteractionCreate, k string) {
		time.Sleep(5 * time.Minute)
		err := cmd.NewResponseBuilder(s, i).IsEdit().Send()
		if err != nil {
			logger.Alert("commands/latex.go - Cannot remove source button", err.Error())
		}
		delete(sourceMap, k)
	}(s, i, k)
}

func OnSourceButton(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionMessageComponent {
		return
	}

	submitData := i.MessageComponentData()
	if submitData.CustomID != GetSourceID {
		logger.Debug("commands/latex.go - not a source button ID")
		return
	}
	resp := cmd.NewResponseBuilder(s, i).IsEphemeral()
	k := fmt.Sprintf("%s:%s", i.ChannelID, i.Message.ID)
	source, ok := sourceMap[k]
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
		renderLatex(s, i, resp, source.StringValue())
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
