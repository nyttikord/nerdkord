package latex

import (
	"fmt"
	"github.com/anhgelus/gokord/utils"
	"github.com/bwmarrin/discordgo"
	"time"
)

var (
	sourceMap = make(map[string]*string, 100)
)

func GetSource(k string) (*string, bool) {
	v, ok := sourceMap[k]
	return v, ok
}

func saveSource(s *discordgo.Session, i *discordgo.InteractionCreate, latexSource string) {
	m, err := s.InteractionResponse(i.Interaction)
	if err != nil {
		utils.SendAlert("commands/latex.go - Getting interaction response", err.Error(), "id", i.ID)
		return
	}
	k := fmt.Sprintf("%s:%s", i.ChannelID, m.ID)
	sourceMap[k] = &latexSource
	utils.SendDebug("source saved", "key", k)
	// remove source button after 5 minutes and clean map
	go func(s *discordgo.Session, i *discordgo.InteractionCreate, k string) {
		time.Sleep(5 * time.Minute)
		err := utils.NewResponseBuilder(s, i).IsEdit().Send()
		if err != nil {
			utils.SendAlert("commands/latex.go - Cannot remove source button", err.Error())
		}
		delete(sourceMap, k)
	}(s, i, k)
}
