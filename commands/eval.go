package commands

import (
	"fmt"
	"github.com/anhgelus/gokord/utils"
	"github.com/bwmarrin/discordgo"
	"strings"
)
import "github.com/nyttikord/gomath"

func Eval(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.Interaction.ApplicationCommandData()
	mathExpr := data.GetOption("expression").StringValue()
	digits := 6

	var precisionOpt = data.GetOption("precision")

	if precisionOpt != nil {
		digits = int(precisionOpt.IntValue())
	}

	precise, err := gomath.ParseAndCalculate(mathExpr, &gomath.Options{Decimal: false})

	resp := utils.NewResponseBuilder(s, i).IsEphemeral()

	if err != nil {
		err = resp.Message("Syntax error: " + err.Error()).
			Send()
		if err != nil {
			utils.SendAlert("commands/eval.go - Sending error", err.Error())
		}
		return
	}

	if digits < 0 {
		err = resp.Message(formatResponse(mathExpr, precise)).
			Send()
		if err != nil {
			utils.SendAlert("commands/eval.go - Sending result", err.Error())
		}
		return
	}
	decimal, _ := gomath.ParseAndCalculate(mathExpr, &gomath.Options{Decimal: true, Precision: int(digits)})

	err = resp.Message(formatResponseDecimal(mathExpr, precise, decimal)).
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
