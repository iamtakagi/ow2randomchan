package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"time"

	"net/http"

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

type HeroDetail struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Portrait    string `json:"portrait"`
	Role        string `json:"role"`
	Location    string `json:"location"`
	Birthday    string `json:"birthday"`
	Age         int    `json:"age"`
	Hitpoints   struct {
		Shields int `json:"shields"`
		Armor   int `json:"armor"`
		Health  int `json:"health"`
		Total   int `json:"total"`
	} `json:"hitpoints"`
	Abilities []struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Icon        string `json:"icon"`
		Video       struct {
			Thumbnail string `json:"thumbnail"`
			Link      struct {
				Mp4  string `json:"mp4"`
				Webm string `json:"webm"`
			} `json:"link"`
		} `json:"video"`
	} `json:"abilities"`
	Story struct {
		Summary string `json:"summary"`
		Media   struct {
			Type string `json:"type"`
			Link string `json:"link"`
		} `json:"media"`
		Chapters []struct {
			Title   string `json:"title"`
			Content string `json:"content"`
			Picture string `json:"picture"`
		} `json:"chapters"`
	} `json:"story"`
}

func fetchHeroDetail(key string) (*HeroDetail, error) {
	response, err := http.Get("https://overfast-api.tekrop.fr/heroes/" + key)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Error: Status code %d", response.StatusCode)
	}

	var heroDetail *HeroDetail

	if err := json.NewDecoder(response.Body).Decode(&heroDetail); err != nil {
		return nil, err
	}

	return heroDetail, nil
}

func main() {
	var err error
	Token = os.Getenv("DISCORD_TOKEN")

	s, err = discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("Discord セッションの作成中にエラーが発生しました,", err)
		return
	}

	raw, err := ioutil.ReadFile("./heroes.json")
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

		detail, err := fetchHeroDetail(hero.Key)
		if err != nil {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "ヒーローの詳細を取得できませんでした。",
				},
			})
			return
		}

		// Assuming detail[0] contains the hero details
		hitpoints := detail.Hitpoints
		embed := &discordgo.MessageEmbed{
			Title:       hero.Name,
			Description: fmt.Sprintf("Role: %s", hero.Role),
			Image: &discordgo.MessageEmbedImage{
				URL: hero.Portrait,
			},
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "Shields",
					Value:  fmt.Sprintf("%d", hitpoints.Shields),
					Inline: true,
				},
				{
					Name:   "Armor",
					Value:  fmt.Sprintf("%d", hitpoints.Armor),
					Inline: true,
				},
				{
					Name:   "Health",
					Value:  fmt.Sprintf("%d", hitpoints.Health),
					Inline: true,
				},
				{
					Name:   "Total",
					Value:  fmt.Sprintf("%d", hitpoints.Total),
					Inline: true,
				},
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
