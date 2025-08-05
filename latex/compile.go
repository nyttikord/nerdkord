package latex

import (
	"bytes"
	"errors"
	"github.com/anhgelus/gokord/cmd"
	"github.com/anhgelus/gokord/logger"
	"github.com/bwmarrin/discordgo"
	"github.com/nyttikord/nerdkord/db"
	"github.com/nyttikord/nerdkord/libs/img"
	"github.com/nyttikord/nerdkord/libs/latex2png"
	"image/color"
	"image/png"
	"math"
	"text/template"
)

var (
	bgColor = color.RGBA{R: 54, G: 57, B: 62, A: 255}
	fgColor = color.White

	defaultPreprocessingOptions = &latex2png.PreprocessingOptions{
		ForbiddenCommands:           []string{"include", "import"},
		CommandsBeforeBeginDocument: []string{"usepackage"},
		TemplateFile:                "config/template.tex",
	}

	defaultPreamble = ""
)

// TestPreamble returns true if the preamble is valid
func TestPreamble(preamble string) bool {
	opt := &*defaultPreprocessingOptions
	opt.UserPreamble = preamble
	return latex2png.Compile(new(bytes.Buffer), "hey, this code was written by a furry", &latex2png.Options{
		LatexBinary:          "latex",
		DvipngBinary:         "dvipng",
		OutputFormat:         latex2png.PNG,
		BackgroundColor:      color.Black,
		ForegroundColor:      color.Black,
		ImageDPI:             10, // reduce DPI for faster results
		PreprocessingOptions: opt,
	}) == nil
}

func RenderLatex(u *discordgo.User, source string) (*bytes.Buffer, error) {
	nerd, err := db.GetNerd(u.ID)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	latexImage, err := png.Decode(file)
	if err != nil {
		return nil, err
	}

	output := new(bytes.Buffer)
	err = png.Encode(output, img.Pad(latexImage, 5+int(math.Ceil(float64(latexImage.Bounds().Dx())*(1./100.))), bgColor))
	if err != nil {
		return nil, err
	}
	return output, nil
}

func RenderLatexAndReply(s *discordgo.Session, i *discordgo.InteractionCreate, resp *cmd.ResponseBuilder, source string, getSourceID string) {
	err := resp.Send()
	if err != nil {
		logger.Alert("latex/compile.go - Sending deferred", err.Error())
		return
	}
	resp.IsEphemeral()

	var u *discordgo.User
	if i.User == nil {
		u = i.Member.User
	} else {
		u = i.User
	}

	output, err := RenderLatex(u, source)

	if err != nil {
		ms, save := handleLatexRenderError(getSourceID, err)
		if save {
			saveSourceWithInteraction(s, i, source)
		}
		resp.IsEphemeral().SetMessage(ms.Content)
		for _, f := range ms.Files {
			resp.AddFile(f)
		}
		for _, c := range ms.Components {
			resp.AddComponent(c)
		}
		if err = resp.Send(); err != nil {
			logger.Alert("latex/compile.go - Sending latex compiling error", err.Error())
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
		logger.Alert("latex/compile.go - Sending latex", err.Error())
		return
	}
	// saving source
	saveSourceWithInteraction(s, i, source)
}

func handleLatexRenderError(getSourceID string, err error) (*discordgo.MessageSend, bool) {
	if errors.As(err, &latex2png.ErrLatexCompilation{}) {
		logger.Debug("latex/compile.go - Latex compilation error")

		msg := &discordgo.MessageSend{}

		if len(err.Error()) > 1950 {
			return &discordgo.MessageSend{
				Content: "⚠️ Compilation error",
				Files: []*discordgo.File{{
					Name:        "error.txt",
					ContentType: "text/plain",
					Reader:      bytes.NewReader([]byte(err.Error())),
				}},
			}, false
		} else {
			msg.Content = "⚠️ Compilation error:\n```\n" + err.Error() + "\n```"
		}
		msg.Components = []discordgo.MessageComponent{discordgo.Button{
			Label:    "",
			Style:    discordgo.SecondaryButton,
			Disabled: false,
			Emoji:    &discordgo.ComponentEmoji{Name: "📝"},
			CustomID: getSourceID,
		}}
		return msg, true
	}
	if errors.Is(err, latex2png.ErrPreprocessor) {
		logger.Debug("latex/compile.go - Preprocessing error")
		return &discordgo.MessageSend{
			Content: "```\n" + err.Error() + "\n```",
		}, false
	}

	logger.Alert("latex/compile.go - Compiling latex", err.Error())
	return &discordgo.MessageSend{
		Content: "Unexpected error while compiling latex. Please report.",
	}, false
}

func GetDefaultPreamble() (string, error) {
	if len(defaultPreamble) == 0 {
		t, err := template.ParseFiles(defaultPreprocessingOptions.TemplateFile)
		if err != nil {
			logger.Alert(
				"commands/preamble.go - Parsing template file", err.Error(),
				"path", defaultPreprocessingOptions.TemplateFile,
			)
		} else {
			wr := new(bytes.Buffer)
			err = t.ExecuteTemplate(wr, "defaultPreamble", nil)
			if err != nil {
				return "", err
			}
			defaultPreamble = wr.String()
		}
	}
	if len(defaultPreamble) == 0 {
		return "Default one", nil
	}
	return defaultPreamble, nil
}
