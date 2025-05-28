package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
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
	Id         int        `json:"id"`
	AwayTeam   GameTeam   `json:"awayTeam"`
	HomeTeam   GameTeam   `json:"homeTeam"`
	StartTime  time.Time  `json:"startTimeUTC"`
	State      string     `json:"gameState"`
	PeriodInfo PeriodDesc `json:"periodDescriptor"`
	DayOfWeek  string
}

type JsonOutput struct {
	AwayAbbrev     string
	HomeAbbrev     string
	StartTimeLocal string
	DayOfWeek      string
}

func NewJsonOutput(aAbbrev string, hAbbrev string, startTime string, dayWeek string) *JsonOutput {
	return &JsonOutput{
		AwayAbbrev:     aAbbrev,
		HomeAbbrev:     hAbbrev,
		StartTimeLocal: startTime,
		DayOfWeek:      dayWeek,
	}
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

type PeriodDesc struct {
	Number int    `json:"number"`
	Type   string `json:"periodType"`
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
	Pre    GameState = "PRE"
	Crit   GameState = "CRIT"
	Final  GameState = "FINAL"
)

const DateFormat = "2006-01-02"
const ShortDayOfWeekFormat = "Mon"

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

	if game.StartTime.Local().Format(DateFormat) == currentDate {
		game.DayOfWeek = "Today"
	} else {
		game.DayOfWeek = game.StartTime.Local().Format(ShortDayOfWeekFormat)
	}

	TextOutput(game)
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
			if len(upperIdentifier) == 3 && upperIdentifier == strings.ToUpper(team.Abbrev) {
				foundTeam = &teams[i]
				return NewGameTeam(foundTeam.Id, foundTeam.Abbrev, 0), nil

			} else if strings.Contains(strings.ToUpper(team.Name), upperIdentifier) {
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

func TextOutput(game *Game) {
	const TimeFormat = "15:04"
	var outputStr string
	var err error
	err = nil
	switch GameState(game.State) {
	case Future:
		outputStr = fmt.Sprintf("%s @ %s | %s @ %s", game.DayOfWeek, game.StartTime.Local().Format(TimeFormat), game.AwayTeam.Abbrev, game.HomeTeam.Abbrev)
	case Live, Pre, Crit:

		var periodStr string
		switch game.PeriodInfo.Number {
		case 1:
			periodStr = strconv.Itoa(game.PeriodInfo.Number) + "st"
		case 2:
			periodStr = strconv.Itoa(game.PeriodInfo.Number) + "nd"
		case 3:
			periodStr = strconv.Itoa(game.PeriodInfo.Number) + "rd"
		default: // Anything higher than 3
			periodStr = "OT"
		}

		outputStr = fmt.Sprintf("%s %d - %s - %d %s", game.AwayTeam.Abbrev, game.AwayTeam.Score, periodStr, game.HomeTeam.Score, game.HomeTeam.Abbrev)
	case Final:
		outputStr = fmt.Sprintf("%s %d - FINAL - %d %s", game.AwayTeam.Abbrev, game.AwayTeam.Score, game.HomeTeam.Score, game.HomeTeam.Abbrev)
	default:
		err = errors.New("Game status unknown")
	}
	if err != nil {
		log.Fatalln("ERROR:", err)
	}

	fmt.Println(outputStr)
}

func MarshalOutput(game *Game) {
	const TimeFormat = "15:04"
	jsonOutput := NewJsonOutput(game.AwayTeam.Abbrev, game.HomeTeam.Abbrev, game.StartTime.Local().Format(TimeFormat), game.DayOfWeek)
	jsonData, err := json.Marshal(jsonOutput)
	if err != nil {
		log.Fatalln("ERROR:", err)
	}
	fmt.Println(string(jsonData))
}
