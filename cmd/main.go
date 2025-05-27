package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

type Team struct {
	Id      int    `json:"id"`
	Name    string `json:"fullName"`
	Tricode string `json:"triCode"`
}
type TeamIdentifier interface {
	~int | ~string
}

type Data struct {
	Data []Team `json:"data"`
}

func main() {
	const TeamEndpt = "https://api.nhle.com/stats/rest/en/team"
	const TeamEndpt = "https://api.nhle.com/stats/rest/en/team"
	const TeamEndpt = "https://api.nhle.com/stats/rest/en/team"

	teamArg, err := GetTeamArg()
	if err != nil {
		log.Fatalln("ERROR:", err)
	}

	teamResp := MakeGetRequest(TeamEndpt)
	teams := GetTeamInfo(teamResp)

	team, err := FindTeam(teamArg, teams)
	if err != nil {
		log.Fatalln("ERROR:", err)
	}
	fmt.Println(team)
}

func MakeGetRequest(url string) *http.Response {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalln("ERROR:", err, "Status Code:", resp.StatusCode)
	}
	return resp
}

func GetTeamInfo(response *http.Response) []Team {
	resp_body := response.Body

	defer resp_body.Close()

	var data Data

	if err := json.NewDecoder(resp_body).Decode(&data); err != nil {
		log.Fatalln("ERROR:", err)
	}

	return data.Data
}

func FindTeam[T TeamIdentifier](teamIdentifier T, teams []Team) (*Team, error) {
	var foundTeam *Team
	for i, team := range teams {
		switch identifier := any(teamIdentifier).(type) {
		case int:
			if identifier == team.Id {
				foundTeam = &teams[i]
				return foundTeam, nil
			}
		case string:
			upperIdentifier := strings.TrimSpace(identifier)
			if strings.Contains(strings.ToUpper(team.Name), upperIdentifier) || upperIdentifier == team.Tricode {
				foundTeam = &teams[i]
				return foundTeam, nil
			}
		default:
			continue
		}
	}
	return foundTeam, errors.New("No team was found with the given tricode.")
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
