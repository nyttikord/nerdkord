package commands

import (
	"github.com/anhgelus/gokord"
	"github.com/anhgelus/gokord/cmd"
	"github.com/anhgelus/gokord/logger"
	"github.com/bwmarrin/discordgo"
)

func About(_ *discordgo.Session, i *discordgo.InteractionCreate, _ cmd.OptionMap, resp *cmd.ResponseBuilder) {
	var u *discordgo.User
	if i.User == nil {
		u = i.Member.User
	} else {
		u = i.User
	}
	if err := resp.SetMessage("**nerdkord**, the open-source Discord bot for nerds made by [Nyttikord](<https://github.com/nyttikord>).\n" +
		"Source code: https://github.com/nyttikord/nerdkord.\n\n" +
		"Host of the bot: " + gokord.BaseCfg.GetAuthor() + ".\n\n" +
		"Uses:\n- [Nyttikord/GoMath](<https://github.com/nyttikord/gomath>)\n" +
		"- [anhgelus/gokord](<https://github.com/anhgelus/gokord>)").Send(); err != nil {
		logger.Alert("commands/about.go - Error while sending about", err.Error(), "discord_id", u.ID)
	}
}
