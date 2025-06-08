package commands

import (
	"fmt"
	"github.com/anhgelus/gokord/utils"
	"github.com/bwmarrin/discordgo"
	"github.com/nyttikord/nerdkord/data"
)

const (
	EditPreambleID = "edit_preamble"
)

func OnProfileButton(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionMessageComponent {
		return
	}
	componentData := i.MessageComponentData()
	if componentData.CustomID != EditPreambleID {
		utils.SendDebug("commands/profile.go - not a profile button ID")
		return
	}
	err := utils.NewResponseBuilder(s, i).
		IsModal().
		SetTitle("Edit preamble").
		SetCustomID(EditPreambleID).
		AddComponent(discordgo.ActionsRow{Components: []discordgo.MessageComponent{
			discordgo.TextInput{
				CustomID:    "source",
				Label:       "Preamble",
				Style:       discordgo.TextInputParagraph,
				Placeholder: `\usepackage[french]{babel}`,
				Required:    true,
				MinLength:   0,
				MaxLength:   4000,
			},
		}}).Send()
	if err != nil {
		utils.SendAlert("profile.go - Sending modal to edit preamble", err.Error(), "discord_id", i.User.ID)
	}
}

func OnProfileModal(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionModalSubmit {
		return
	}
	submitData := i.ModalSubmitData()
	if submitData.CustomID != EditPreambleID {
		utils.SendDebug("commands/profile.go - not a profile modal ID")
		return
	}
	resp := utils.NewResponseBuilder(s, i).IsEphemeral()
	nerd, err := data.GetNerd(i.User.ID)
	if err != nil {
		utils.SendAlert("commands/latex.go - Getting nerd", err.Error(), "discord_id", i.User.ID)
		if err = resp.SetMessage("Error while getting your profile. Please report the bug.").Send(); err != nil {
			utils.SendAlert("commands/latex.go - Sending error getting nerd", err.Error())
		}
		return
	}

	nerd.Preamble = submitData.Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
	err = nerd.Save()
	if err != nil {
		utils.SendAlert("commands/latex.go - Saving preamble", err.Error(), "discord_id", i.User.ID)
		if err = resp.SetMessage("Error while saving your profile. Please report the bug.").Send(); err != nil {
			utils.SendAlert("commands/latex.go - Sending error getting nerd", err.Error())
		}
		return
	}
	if err = resp.SetMessage("Preamble saved").Send(); err != nil {
		utils.SendAlert("commands/latex.go - Sending success", err.Error())
	}
}

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
	}).AddComponent(discordgo.ActionsRow{Components: []discordgo.MessageComponent{
		discordgo.Button{
			Label:    "Edit preamble",
			Style:    discordgo.PrimaryButton,
			Disabled: false,
			Emoji:    nil,
			CustomID: EditPreambleID,
		},
	},
	}).Send()
	if err != nil {
		utils.SendAlert("profile.go - Sending profile", err.Error(), "discord_id", i.User.ID)
	}
}
