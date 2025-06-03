package commands

import (
	"fmt"
	"github.com/anhgelus/gokord/utils"
	"github.com/bwmarrin/discordgo"
	"math"
	"strings"
)
import "github.com/nyttikord/gomath"

func Eval(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.Interaction.ApplicationCommandData()
	mathExpr := data.GetOption("expression").StringValue()
	digits := 6

	var precisionOpt = data.GetOption("precision")

	if precisionOpt != nil {
		digits = int(math.Abs(float64(precisionOpt.IntValue())))
	}

	precise, err := gomath.ParseAndCalculate(mathExpr, &gomath.Options{Decimal: false})

	resp := utils.NewResponseBuilder(s, i).IsEphemeral()

	if err != nil {
		resp.Message("Syntax error: " + err.Error())
		err = resp.Send()
		if err != nil {
			utils.SendAlert("commands/eval.go - Sending error", err.Error())
		}
		return
	}

	decimal, _ := gomath.ParseAndCalculate(mathExpr, &gomath.Options{Decimal: true, Precision: int(digits)})

	resp.Message(formatResponse(mathExpr, precise, decimal))

	err = resp.Send()
	if err != nil {
		utils.SendAlert("commands/eval.go - Sending result", err.Error())
	}
}

func formatResponse(expr string, precise string, decimal string) string {
	return fmt.Sprintf("```\n"+
		"%s = %s"+
		"\n"+strings.Repeat(" ", len(expr))+" ≈ %s"+
		"\n```", expr, precise, decimal)
}
