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

func saveSourceWithInteraction(s *discordgo.Session, i *discordgo.InteractionCreate, source string) {
	m, err := s.InteractionResponse(i.Interaction)
	if err != nil {
		utils.SendAlert("commands/latex.go - Getting interaction response", err.Error(), "id", i.ID)
		return
	}
	saveSource(s, m, source, func(s *discordgo.Session) {
		err := utils.NewResponseBuilder(s, i).IsEdit().Send()
		if err != nil {
			utils.SendAlert("commands/latex.go - Cannot remove source button", err.Error())
		}
	})
}

func saveSourceWithMessage(s *discordgo.Session, m *discordgo.Message, source string) {
	saveSource(s, m, source, func(s *discordgo.Session) {
		_, err := s.ChannelMessageEditComplex(&discordgo.MessageEdit{
			Content:     &m.Content,
			Components:  nil,
			Attachments: &m.Attachments,
			ID:          m.ID,
			Channel:     m.ChannelID,
		})
		if err != nil {
			utils.SendAlert("commands/latex.go - Cannot remove source button", err.Error())
		}
	})
}

func saveSource(s *discordgo.Session, m *discordgo.Message, source string, fn func(s *discordgo.Session)) {
	k := fmt.Sprintf("%s:%s", m.ChannelID, m.ID)
	sourceMap[k] = &source
	utils.SendDebug("source saved", "key", k)
	// remove source button after 5 minutes and clean map
	go func(s *discordgo.Session, k string) {
		time.Sleep(5 * time.Minute)
		fn(s)
		delete(sourceMap, k)
	}(s, k)
}
