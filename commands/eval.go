package commands

import (
	"fmt"
	"github.com/anhgelus/gokord/utils"
	"github.com/bwmarrin/discordgo"
	"strings"
)
import "github.com/nyttikord/gomath"

func Eval(_ *discordgo.Session, _ *discordgo.InteractionCreate, optMap utils.OptionMap, resp *utils.ResponseBuilder) {
	mathExprOpt, ok := optMap["expression"]
	resp.IsEphemeral()

	if !ok {
		utils.SendAlert("commands/eval.go - Getting expression option", "expression option is not present")

		err := resp.
			SetMessage("An error occurred while running this command. Try again later, or contact a bot developer").
			Send()

		if err != nil {
			utils.SendAlert("commands/eval.go - Sending internal error message", err.Error())
		}

		return
	}
	mathExpr := mathExprOpt.StringValue()

	digits := 6
	precisionOpt, ok := optMap["precision"]
	if ok {
		digits = int(precisionOpt.IntValue())
	}

	precise, err := gomath.ParseAndCalculate(mathExpr, &gomath.Options{Decimal: false})

	if err != nil {
		err = resp.SetMessage("Syntax error: " + err.Error()).
			Send()
		if err != nil {
			utils.SendAlert("commands/eval.go - Sending error", err.Error())
		}
		return
	}

	if digits < 0 {
		err = resp.SetMessage(formatResponse(mathExpr, precise)).
			Send()
		if err != nil {
			utils.SendAlert("commands/eval.go - Sending result", err.Error())
		}
		return
	}
	decimal, _ := gomath.ParseAndCalculate(mathExpr, &gomath.Options{Decimal: true, Precision: int(digits)})

	err = resp.SetMessage(formatResponseDecimal(mathExpr, precise, decimal)).
		Send()
	if err != nil {
		utils.SendAlert("commands/eval.go - Sending decimal result", err.Error())
	}
}

func formatResponse(expr string, precise string) string {
	return fmt.Sprintf("```\n"+
		"%s = %s"+
		"\n```", expr, precise)
}

func formatResponseDecimal(expr string, precise string, decimal string) string {
	return fmt.Sprintf("```\n"+
		"%s = %s"+
		"\n"+strings.Repeat(" ", len(expr))+" ≈ %s"+
		"\n```", expr, precise, decimal)
}
