package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"time"

	"github.com/bwmarrin/discordgo"
)

var (
	Token string
	s     *discordgo.Session
	heros []Hero
)

type Hero struct {
	Key      string `json:"key"`
	Name     string `json:"name"`
	Portrait string `json:"portrait"`
	Role     string `json:"role"`
}

func main() {
	var err error
	Token = os.Getenv("DISCORD_TOKEN")

	s, err = discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("Discord セッションの作成中にエラーが発生しました,", err)
		return
	}

	raw, err := ioutil.ReadFile("./heros.json")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	err = json.Unmarshal(raw, &heros)
	if err != nil {
		fmt.Println("JSON のアンマーシャル中にエラーが発生しました:", err)
		os.Exit(1)
	}

	err = s.Open()
	if err != nil {
		fmt.Println("セッションのオープン中にエラーが発生しました:", err)
		return
	}

	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "r",
			Description: "ランダムにヒーローをピックします",
		},
	}

	for _, g := range s.State.Guilds {
		_, err := s.ApplicationCommandCreate(s.State.User.ID, g.ID, commands[0])
		if err != nil {
			fmt.Printf("Cannot create '%v' command: %v\n", commands[0].Name, err)
		}
	}

	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})

	fmt.Println("ボットが起動しました。CTRL+C で終了します。")
	select {}
}

var commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	"r": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		hero := pickHeroRandom()
		embed := &discordgo.MessageEmbed{
			Title:       hero.Name,
			Description: fmt.Sprintf("役割: %s", hero.Role),
			Image: &discordgo.MessageEmbedImage{
				URL: hero.Portrait,
			},
		}

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{embed},
			},
		})
	},
}

func pickHeroRandom() Hero {
	rand.Seed(time.Now().UnixNano())
	randomIndex := rand.Intn(len(heros))
	return heros[randomIndex]
}
