package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type TeamInfo struct {
	Id     int    `json:"id"`
	Name   string `json:"fullName"`
	Abbrev string `json:"triCode"`
}

type TeamIdentifier interface {
	~int | ~string
}

type Data struct {
	Data []TeamInfo `json:"data"`
}

// Make getter and setter for the time to be converted to local time
type Game struct {
	Id        int       `json:"id"`
	AwayTeam  GameTeam  `json:"awayTeam"`
	HomeTeam  GameTeam  `json:"homeTeam"`
	StartTime time.Time `json:"startTimeUTC"`
	GameState string    `json:"gameState"`
}

type GameTeam struct {
	Id     int    `json:"id"`
	Abbrev string `json:"abbrev"`
	Score  int    `json:"score"`
}

func NewGameTeam(id int, abbrev string, score int) *GameTeam {
	return &GameTeam{
		Id:     id,
		Abbrev: abbrev,
		Score:  score,
	}
}

type GameDay struct {
	DateStr   string `json:"date"`
	DayAbbrev string `json:"dayAbbrev"`
	Games     []Game `json:"games"`
}

type GameWeek struct {
	GameDays []GameDay `json:"gameWeek"`
}

type GameState string

const (
	Future GameState = "FUT"
	Off    GameState = "OFF"
	Live   GameState = "LIVE"
)

const DateFormat = "2006-01-02"

func main() {
	const TeamEndpt = "https://api.nhle.com/stats/rest/en/team"

	// This needs the current date attached to the end
	// YYYY-MM-DD format
	const ScheduleNowEndpt = "https://api-web.nhle.com/v1/schedule/"

	teamArg, err := GetTeamArg()
	if err != nil {
		log.Fatalln("ERROR:", err)
	}

	teamResp := MakeGetRequest(TeamEndpt)
	teams := GetTeamInfo(teamResp)

	chosenTeam, err := FindTeam(teamArg, teams)
	if err != nil {
		log.Fatalln("ERROR:", err)
	}

	currentDate := time.Now().Format(DateFormat)
	schedNowResp := MakeGetRequest(ScheduleNowEndpt + currentDate)
	gameWeek := GetGamesThisWeek(schedNowResp)
	game, found := gameWeek.GetTeamNextGame(chosenTeam)
	if !found {
		fmt.Println("No games scheduled.")
		return
	}
	fmt.Println(game)
}

func MakeGetRequest(url string) *http.Response {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalln("ERROR:", err, "Status Code:", resp.StatusCode)
	}
	return resp
}

func GetTeamInfo(response *http.Response) []TeamInfo {
	resp_body := response.Body

	defer resp_body.Close()

	var data Data

	if err := json.NewDecoder(resp_body).Decode(&data); err != nil {
		log.Fatalln("ERROR:", err)
	}

	return data.Data
}

func FindTeam[T TeamIdentifier](teamIdentifier T, teams []TeamInfo) (*GameTeam, error) {
	var foundTeam *TeamInfo
	for i, team := range teams {
		switch identifier := any(teamIdentifier).(type) {
		case int:
			if identifier == team.Id {
				foundTeam = &teams[i]
				return NewGameTeam(foundTeam.Id, foundTeam.Abbrev, 0), nil
			}
		case string:
			upperIdentifier := strings.TrimSpace(identifier)
			if strings.Contains(strings.ToUpper(team.Name), upperIdentifier) {
				foundTeam = &teams[i]
				return NewGameTeam(foundTeam.Id, foundTeam.Abbrev, 0), nil
			}
		default:
			continue
		}
	}
	return nil, errors.New("No team was found with the given identifier.")
}

func GetTeamArg() (string, error) {
	args := os.Args
	var teamArgUpper string
	var err error = nil

	argNum := len(args)
	if argNum < 2 {
		err = errors.New("No argument passed.")
		return teamArgUpper, err
	} else if argNum > 2 {
		err = errors.New("Too may arguments passed.")
		return teamArgUpper, err
	}

	teamArgUpper = strings.ToUpper(args[1])
	return teamArgUpper, err
}

func GetGamesThisWeek(response *http.Response) GameWeek {
	resp_body := response.Body
	defer resp_body.Close()

	var gameWeek GameWeek
	if err := json.NewDecoder(resp_body).Decode(&gameWeek); err != nil {
		log.Fatalln("ERROR:", err)
	}
	return gameWeek
}

func (gW *GameWeek) GetTeamNextGame(chosenTeamPt *GameTeam) (*Game, bool) {
	for _, gameDay := range gW.GameDays {
		for _, game := range gameDay.Games {
			if game.AwayTeam.Abbrev == chosenTeamPt.Abbrev || game.HomeTeam.Abbrev == chosenTeamPt.Abbrev {
				return &game, true
			}
		}
	}
	return nil, false
}
