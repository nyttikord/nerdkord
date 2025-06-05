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

	res, err := gomath.ParseAndConvertToLatex(expr, &gomath.Options{})

	if err != nil {
		resp.SetMessage("Syntax error: " + err.Error())
		err = resp.Send()
		if err != nil {
			utils.SendAlert("commands/latexify.go - Sending error", err.Error())
		}
		return
	}

	err = resp.SetMessage(fmt.Sprintf("LaTeX code of `%s`: \n```\n%s\n```", expr, res)).
		Send()

	if err != nil {
		utils.SendAlert("commands/latexify.go - Sending result", err.Error())
	}
}
