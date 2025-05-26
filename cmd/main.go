package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
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

//func NewTeam(name, tricode string, id int) *Team {
//	return &Team{
//		Id:      id,
//		Name:    name,
//		Tricode: tricode,
//	}
//}

func main() {
	// Get all the team information
	teamResp := MakeGetRequest("https://api.nhle.com/stats/rest/en/team")
	teams := GetTeamInfo(teamResp)
	team, _ := FindTeam("	oilers", teams)
	fmt.Println(team.Id, team.Name, team.Tricode)
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
			upperIdentifier := strings.ToUpper(strings.TrimSpace(identifier))
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
