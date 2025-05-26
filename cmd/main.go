package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type Team struct {
	Id      int    `json:"id"`
	Name    string `json:"fullName"`
	Tricode string `json:"triCode"`
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
	for _, team := range teams {
		fmt.Println(team.Name, team.Tricode)
	}

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
