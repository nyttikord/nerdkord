package commands

import (
	"fmt"
	"github.com/anhgelus/gokord/utils"
	"github.com/bwmarrin/discordgo"
	"github.com/nyttikord/nerdkord/data"
)

func Profile(_ *discordgo.Session, i *discordgo.InteractionCreate, optMap utils.OptionMap, resp *utils.ResponseBuilder) {
	resp.IsEphemeral()
	nerd, err := data.GetNerd(i.User.ID)
	if err != nil {
		utils.SendAlert("profile.go - Getting nerd", err.Error(), "discord_id", i.User.ID)
		if err = resp.SetMessage("Error while getting your profile. Please report.").Send(); err != nil {
			utils.SendAlert("profile.go - Getting nerd error", err.Error(), "discord_id", i.User.ID)
		}
		return
	}
	if len(nerd.Preamble) == 0 {
		nerd.Preamble = "Default one"
	}
	err = resp.AddEmbed(&discordgo.MessageEmbed{
		Title:       fmt.Sprintf("%s's nerd profile", i.User.Username),
		Description: fmt.Sprintf("Your preamble:\n```tex\n%s\n```", nerd.Preamble),
		Color:       0,
	}).Send()
	if err != nil {
		utils.SendAlert("profile.go - Sending profile", err.Error(), "discord_id", i.User.ID)
	}
}
