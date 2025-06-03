package commands

import (
	"fmt"
	"github.com/anhgelus/gokord/utils"
	"github.com/bwmarrin/discordgo"
	"github.com/nyttikord/gomath"
)

func Latexify(s *discordgo.Session, i *discordgo.InteractionCreate) {
	resp := utils.NewResponseBuilder(s, i).IsEphemeral()

	optMap := utils.GenerateOptionMap(i)
	exprOpt, ok := optMap["expression"]

	if !ok {
		utils.SendAlert("commands/latexify.go - Getting expression option", "expression option is not present")

		err := utils.NewResponseBuilder(s, i).
			IsEphemeral().
			Message("An error occurred while running this command. Try again later, or contact a bot developer").
			Send()

		if err != nil {
			utils.SendAlert("commands/latexify.go - Sending internal error message", err.Error())
		}

		return
	}
	expr := exprOpt.StringValue()

	res, err := gomath.ParseAndConvertToLatex(expr, &gomath.Options{})

	if err != nil {
		resp.Message("Syntax error: " + err.Error())
		err = resp.Send()
		if err != nil {
			utils.SendAlert("commands/latexify.go - Sending error", err.Error())
		}
		return
	}

	err = resp.Message(fmt.Sprintf("LaTeX code of `%s`: \n```\n%s\n```", expr, res)).
		Send()

	if err != nil {
		utils.SendAlert("commands/latexify.go - Sending result", err.Error())
	}
}
