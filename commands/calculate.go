package commands

import (
	"fmt"
	"github.com/anhgelus/gokord/utils"
	"github.com/bwmarrin/discordgo"
	"strings"
)
import "github.com/nyttikord/gomath"

func Calculate(s *discordgo.Session, i *discordgo.InteractionCreate, optMap utils.OptionMap, resp *utils.ResponseBuilder) {
	mathExprOpt, ok := optMap["expression"]
	resp.IsEphemeral()

	if !ok {
		utils.SendAlert("commands/calculate.go - Getting expression option", "expression option is not present")

		err := resp.
			SetMessage("An error occurred while running this command. Try again later, or contact a bot developer").
			Send()

		if err != nil {
			utils.SendAlert("commands/calculate.go - Sending internal error message", err.Error())
		}

		return
	}
	mathExpr := mathExprOpt.StringValue()

	digits := 6
	precisionOpt, ok := optMap["precision"]
	if ok {
		digits = int(precisionOpt.IntValue())
	}

	result, err := gomath.Parse(mathExpr)

	if err != nil {
		err = resp.SetMessage("Syntax error: " + err.Error()).
			Send()
		if err != nil {
			utils.SendAlert("commands/calculate.go - Sending error", err.Error())
		}
		return
	}

	err = resp.SetMessage(formatResponse(mathExpr, result, digits)).
		Send()
	if err != nil {
		utils.SendAlert("commands/calculate.go - Sending decimal result", err.Error())
	}
}

func formatResponse(expr string, result gomath.Result, precision int) string {
	if precision < -1 || result.IsExact(precision) {
		return fmt.Sprintf("```\n%s = %s\n```", expr, result.String())
	}

	return fmt.Sprintf("```\n"+
		"%s = %s"+
		"\n"+strings.Repeat(" ", len(expr))+" ≈ %s"+
		"\n```", expr, result.String(), result.Approx(precision))
}
