package latex

import (
	"bytes"
	"errors"
	"github.com/anhgelus/gokord/utils"
	"github.com/bwmarrin/discordgo"
	"github.com/nyttikord/nerdkord/db"
	"github.com/nyttikord/nerdkord/libs/img"
	"github.com/nyttikord/nerdkord/libs/latex2png"
	"image/color"
	"image/png"
	"math"
)

var (
	bgColor = color.RGBA{R: 54, G: 57, B: 62, A: 255}
	fgColor = color.White

	defaultPreprocessingOptions = &latex2png.PreprocessingOptions{
		ForbiddenCommands:           []string{"include", "import"},
		CommandsBeforeBeginDocument: []string{"usepackage"},
		TemplateFile:                "config/template.tex",
	}
)

func RenderLatex(s *discordgo.Session, i *discordgo.InteractionCreate, resp *utils.ResponseBuilder, source string, getSourceID string) {
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

	nerd, err := db.GetNerd(u.ID)
	if err != nil {
		utils.SendAlert("commands/latex.go - Getting nerd", err.Error(), "discord_id", u.ID)
		if err = resp.SetMessage("Error while getting your profile. Please report the bug.").Send(); err != nil {
			utils.SendAlert("commands/latex.go - Sending error getting nerd", err.Error())
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
		handleLatexRenderError(s, i, resp, source, getSourceID, err)
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
			Label:    "",
			Style:    discordgo.SecondaryButton,
			Disabled: false,
			Emoji:    &discordgo.ComponentEmoji{Name: "📝"},
			CustomID: getSourceID,
		},
	}}).Send()
	if err != nil {
		utils.SendAlert("commands/latex.go - Sending latex", err.Error())
		return
	}
	// saving source
	saveSource(s, i, source)
}

func handleLatexRenderError(s *discordgo.Session, i *discordgo.InteractionCreate, resp *utils.ResponseBuilder, source string, getSourceID string, err error) {
	if errors.As(err, &latex2png.ErrLatexCompilation{}) {
		utils.SendDebug("commands/latex.go - Latex compilation error")

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
				CustomID: getSourceID,
			},
		}}).Send()
		if err != nil {
			utils.SendAlert("commands/latex.go - Sending compilation error", err.Error())
		}

		saveSource(s, i, source)
		return
	}
	if errors.Is(err, latex2png.ErrPreprocessor) {
		utils.SendDebug("commands/latex.go - Preprocessing error")
		err = resp.SetMessage("```\n" + err.Error() + "\n```").Send()
		if err != nil {
			utils.SendAlert("commands/latex.go - Sending preprocessing error", err.Error())
		}
		return
	}

	utils.SendAlert("commands/latex.go - Compiling latex", err.Error())
	err = resp.SetMessage("Unexpected error while compiling latex").Send()
	if err != nil {
		utils.SendAlert("commands/latex.go - Sending unexpected latex error", err.Error())
	}
}
