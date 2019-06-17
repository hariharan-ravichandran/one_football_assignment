package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strconv"
)

// player represents a player data from the fetchedData
type player struct {
	Name string
	Age  string
}

// players represents a group of players
type players []player

// fetchedData represents the data decoded from the fetched JSON
type fetchedData struct {
	Status string
	Data   struct {
		Team struct {
			IsNational bool
			Name       string
			Players    players
		}
	}
}

// playerInfo represents the age and teams of the players to be rendered in the output
type playerInfo struct {
	age   string
	teams []string
}

var neededTeams = []string{"Germany", "England", "France", "Spain", "Manchester United", "Arsenal", "Chelsea", "Barcelona", "Real Madrid", "Bayern Munich"}

func main() {
	var (
		playerNames []string
		loopCount   int
	)
	playerInfos := make(map[string]playerInfo)

	client := &http.Client{}

	for {
		loopCount += 1
		team_id := strconv.Itoa(loopCount) + ".json"
		url := fmt.Sprintf("https://vintagemonster.onefootball.com/api/teams/en/%s", team_id)
		var output fetchedData
		if dataObtained := fetchData(client, url, &output); !dataObtained {
			continue
		}
		playerNames = populateDataForOutput(output, playerNames, playerInfos)
		if len(neededTeams) == 0 {
			break
		}
	}
	renderOutput(playerNames, playerInfos)
}

// renderOutput is used to send the player information to the stdOut in alphabetical order.
func renderOutput(playerNames []string, playerInfos map[string]playerInfo) {
	sort.Strings(playerNames)
	for _, playerName := range playerNames {
		playerInf := playerInfos[playerName]
		chunk := "\n" + playerName + "; " + playerInf.age + "; "
		for index, team := range playerInf.teams {
			if index != 0 {
				chunk += ", "
			}
			chunk += team
		}
		os.Stdout.Write([]byte(chunk))
	}
}

// populateDataForOutput is used to populate the player information from the decoded Go structure.
func populateDataForOutput(output fetchedData, playerNames []string, playerInfos map[string]playerInfo) []string {
	if output.Status == "ok" {
		var found bool
		if found, neededTeams = findAndRemoveString(neededTeams, output.Data.Team.Name); found {
			for _, player := range output.Data.Team.Players {
				if existingInfo, exists := playerInfos[player.Name]; !exists {
					playerInfos[player.Name] = playerInfo{age: player.Age, teams: []string{output.Data.Team.Name}}
					playerNames = append(playerNames, player.Name)
				} else {
					playerInfos[player.Name] = playerInfo{age: existingInfo.age, teams: append(existingInfo.teams, output.Data.Team.Name)}
				}
			}
		}
	}
	return playerNames
}

// fetchData fetches data from the respective URL and populates the given struct and returns a bool
// indicating whether the struct is populated or not.
func fetchData(client *http.Client, url string, output *fetchedData) bool {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		os.Stderr.WriteString(err.Error())
		return false
	}
	response, err := client.Do(req)
	if err != nil {
		os.Stderr.WriteString(err.Error())
		return false
	}
	if response.StatusCode != 200 {
		return false
	}
	defer response.Body.Close()
	if err = json.NewDecoder(response.Body).Decode(&output); err != nil {
		os.Stderr.WriteString(err.Error())
		return false
	}
	return true
}

// findAndRemove gets an array of string and a particular string. It searches the particular string
// in the array and removes that element from the array and returns two values.
// 1 - A bool representing whether that particular string was present.
// 2 - An arrat with removed element if that is present, else the input array is returned.
func findAndRemoveString(list []string, word string) (bool, []string) {
	for index, w := range list {
		if w == word {
			return true, removeElement(list, index)
		}
	}
	return false, list
}

// removeElement removes the data at a particular index from the given array of string
// and returns the modified array.
func removeElement(list []string, index int) []string {
	list[len(list)-1], list[index] = list[index], list[len(list)-1]
	return list[:len(list)-1]
}
