package commands

import (
	"fmt"
	"github.com/anhgelus/gokord/utils"
	"github.com/bwmarrin/discordgo"
	"github.com/nyttikord/gomath"
)

func Latexify(_ *discordgo.Session, _ *discordgo.InteractionCreate, optMap utils.OptionMap, resp *utils.ResponseBuilder) {
	resp.IsEphemeral()

	exprOpt, ok := optMap["expression"]

	if !ok {
		utils.SendAlert("commands/latexify.go - Getting expression option", "expression option is not present")

		err := resp.
			SetMessage("An error occurred while running this command. Try again later, or contact a bot developer").
			Send()

		if err != nil {
			utils.SendAlert("commands/latexify.go - Sending internal error message", err.Error())
		}

		return
	}
	expr := exprOpt.StringValue()

	result, err := gomath.Parse(expr)

	if err != nil {
		err = resp.SetMessage("Syntax error: " + err.Error()).Send()
		if err != nil {
			utils.SendAlert("commands/latexify.go - Sending syntax error", err.Error())
		}
		return
	}

	latex, err := result.LaTeX()

	if err != nil {
		utils.SendDebug("commands/latexify.go - Couldn't convert to latex")
		err = resp.SetMessage("Couldn't convert expression to LaTeX.").Send()
		if err != nil {
			utils.SendAlert("commands/latexify.go - Sending latex conversion error", err.Error())
		}
	}

	err = resp.SetMessage(fmt.Sprintf("LaTeX code of `%s`: \n```latex\n%s\n```", expr, latex)).
		Send()

	if err != nil {
		utils.SendAlert("commands/latexify.go - Sending result", err.Error())
	}
}
