package commands

import (
	"fmt"
	"github.com/anhgelus/gokord/cmd"
	"github.com/anhgelus/gokord/logger"
	"github.com/bwmarrin/discordgo"
	"github.com/nyttikord/nerdkord/db"
	"github.com/nyttikord/nerdkord/latex"
	"strings"
)

const (
	EditPreambleID        = "edit_preamble"
	ResetPreambleID       = "reset_preamble"
	ReallyResetPreambleID = "really_reset_preamble"
)

func OnEditPreambleButton(_ *discordgo.Session, i *discordgo.InteractionCreate, _ discordgo.MessageComponentInteractionData, resp *cmd.ResponseBuilder) {
	var u *discordgo.User
	if i.User != nil {
		u = i.User
	} else {
		u = i.Member.User
	}
	var val string
	nerd, err := db.GetNerd(u.ID)
	if err != nil {
		logger.Warn("Getting nerd profile", "err", err.Error(), "discord_id", u.ID)

		val, err = latex.GetDefaultPreamble()
		if err == nil {
			logger.Debug("Using default preamble as placeholder")
		} else {
			logger.Alert("commands/preamble.go - Getting default preamble", err.Error())
			logger.Debug("Using empty preamble")
			val = ""
		}
	} else {
		val = nerd.Preamble
		if len(nerd.Preamble) == 0 {
			val, err = latex.GetDefaultPreamble()
			if err != nil {
				logger.Warn("Getting default preamble", "err", err.Error())
			}
		}
	}
	err = resp.IsModal().
		SetTitle("Edit preamble").
		SetCustomID(EditPreambleID).
		AddComponent(discordgo.ActionsRow{Components: []discordgo.MessageComponent{
			discordgo.TextInput{
				CustomID:    "source",
				Label:       "Preamble",
				Style:       discordgo.TextInputParagraph,
				Placeholder: `\usepackage[french]{babel}`,
				Value:       val,
				Required:    true,
				MinLength:   0,
				MaxLength:   4000,
			},
		}}).Send()
	if err != nil {
		logger.Alert("commands/preamble.go - Sending modal to edit preamble", err.Error(), "discord_id", u.ID)
	}
}

func OnResetPromptPreambleButton(_ *discordgo.Session, _ *discordgo.InteractionCreate, _ discordgo.MessageComponentInteractionData, resp *cmd.ResponseBuilder) {
	err := resp.IsEphemeral().
		SetMessage("Are you sure you want to reset your preamble ?").
		AddComponent(&discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Yes, reset my preamble",
					Style:    discordgo.DangerButton,
					Emoji:    &discordgo.ComponentEmoji{Name: "⚠️"},
					Disabled: false,
					CustomID: ReallyResetPreambleID,
				},
			},
		}).Send()

	if err != nil {
		logger.Alert("commands/preamble.go - Sending reset confirmation", err.Error())
	}
}

func OnReallyResetPromptPreambleButton(_ *discordgo.Session, i *discordgo.InteractionCreate, _ discordgo.MessageComponentInteractionData, resp *cmd.ResponseBuilder) {
	resp.IsEphemeral()
	var u *discordgo.User
	if i.User == nil {
		u = i.Member.User
	} else {
		u = i.User
	}
	nerd, err := db.GetNerd(u.ID)
	if err != nil {
		logger.Alert("commands/preamble.go - Getting nerd", err.Error(), "discord_id", u.ID)
		if err = resp.SetMessage("Error while getting your preamble. Please report the bug.").Send(); err != nil {
			logger.Alert("commands/preamble.go - Sending error getting nerd", err.Error())
		}
		return
	}

	nerd.Preamble = ""

	err = nerd.Save()
	if err != nil {
		logger.Alert("commands/preamble.go - Resetting preamble", err.Error(), "discord_id", u.ID)
		if err = resp.SetMessage("Error while resetting your preamble. Please report the bug.").Send(); err != nil {
			logger.Alert("commands/preamble.go - Sending error resetting preamble", err.Error())
		}
		return
	}
	if err = resp.SetMessage("Preamble reset to default").Send(); err != nil {
		logger.Alert("commands/preamble.go - Sending reset success", err.Error())
	}
}

func OnPreambleModalSubmit(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionModalSubmit {
		return
	}
	submitData := i.ModalSubmitData()
	if submitData.CustomID != EditPreambleID {
		logger.Debug("commands/preamble.go - not a profile modal ID")
		return
	}
	resp := cmd.NewResponseBuilder(s, i).IsEphemeral()
	var u *discordgo.User
	if i.User == nil {
		u = i.Member.User
	} else {
		u = i.User
	}
	nerd, err := db.GetNerd(u.ID)
	if err != nil {
		logger.Alert("commands/preamble.go - Getting nerd", err.Error(), "discord_id", u.ID)
		if err = resp.SetMessage("Error while getting your preamble. Please report the bug.").Send(); err != nil {
			logger.Alert("commands/preamble.go - Sending error getting nerd", err.Error())
		}
		return
	}

	val := submitData.Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
	if strings.Contains(val, `\documentclass`) {
		if err = resp.SetMessage("You can't use `\\documentclass`").Send(); err != nil {
			logger.Alert("commands/preamble.go - Sending error \\documentclass is present", err.Error())
		}
		return
	}
	if err = resp.IsDeferred().Send(); err != nil {
		logger.Alert("commands/preamble.go - Sending deferred", err.Error(), "discord_id", u.ID)
		return
	}
	logger.Debug("Checking preamble's validity", "discord_id", u.ID)
	if !latex.TestPreamble(val) {
		if err = resp.SetMessage("Your preamble is invalid.").Send(); err != nil {
			logger.Alert("commands/preamble.go - Sending invalid preamble", err.Error())
		}
		return
	}
	logger.Debug("Updating nerd's preamble", "discord_id", u.ID)
	nerd.Preamble = val
	err = nerd.Save()
	if err != nil {
		logger.Alert("commands/preamble.go - Saving preamble", err.Error(), "discord_id", u.ID)
		if err = resp.SetMessage("Error while saving your preamble. Please report the bug.").Send(); err != nil {
			logger.Alert("commands/preamble.go - Sending error getting nerd", err.Error())
		}
		return
	}
	if err = resp.SetMessage("Preamble saved").Send(); err != nil {
		logger.Alert("commands/preamble.go - Sending success", err.Error())
	}
}

func Preamble(_ *discordgo.Session, i *discordgo.InteractionCreate, _ cmd.OptionMap, resp *cmd.ResponseBuilder) {
	resp.IsEphemeral()
	var u *discordgo.User
	if i.User == nil {
		u = i.Member.User
	} else {
		u = i.User
	}
	nerd, err := db.GetNerd(u.ID)
	if err != nil {
		logger.Alert("commands/preamble.go - Getting nerd", err.Error(), "discord_id", u.ID)
		if err = resp.SetMessage("Error while getting your preamble. Please report.").Send(); err != nil {
			logger.Alert("commands/preamble.go - Getting nerd error", err.Error(), "discord_id", u.ID)
		}
		return
	}
	if len(nerd.Preamble) == 0 {
		nerd.Preamble, err = latex.GetDefaultPreamble()
		if err != nil {
			if err = resp.SetMessage("An error occurred. Please report the bug.").Send(); err != nil {
				logger.Alert("commands/preamble.go - Sending error occurred while parsing template", err.Error())
			}
			return
		}
	}
	err = resp.AddEmbed(&discordgo.MessageEmbed{
		Title:       fmt.Sprintf("%s's preamble", u.Username),
		Description: fmt.Sprintf("Your preamble:\n```tex\n%s\n```", nerd.Preamble),
		Color:       0,
	}).AddComponent(discordgo.ActionsRow{Components: []discordgo.MessageComponent{
		discordgo.Button{
			Label:    "Edit",
			Style:    discordgo.SecondaryButton,
			Disabled: false,
			Emoji:    &discordgo.ComponentEmoji{Name: "✏️"},
			CustomID: EditPreambleID,
		},
		discordgo.Button{
			Label:    "Reset",
			Style:    discordgo.SecondaryButton,
			Disabled: false,
			Emoji:    &discordgo.ComponentEmoji{Name: "🔄"},
			CustomID: ResetPreambleID,
		},
	}}).Send()
	if err != nil {
		logger.Alert("commands/preamble.go - Sending preamble", err.Error(), "discord_id", u.ID)
	}
}
