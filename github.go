package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Actor struct {
	ID          int    `json:"id"`
	DisplayName string `json:"display_login"`
}

type Repo struct {
	ID   int    `json="id"`
	Name string `json:"name"`
	Url  string `json="url"`
}

type Event struct {
	ID    string `json:"id"`
	Type  string `json:"type"`
	Actor Actor  `json:"actor"`
	Repo  Repo   `json:"repo"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	username := os.Args[1]
	token := os.Getenv("GHTOKEN")

	events := callEndpoint(token, username)
	organisedEvents := getEventsForEachRepo(events)

	outputs := getOutputs(organisedEvents)
	fmt.Println("Outputs:")
	for _, output := range outputs {
		fmt.Println(output)
	}
}

func getOutputs(events map[string]map[string]int) []string {
	var outputs []string

	for eventType, value := range events {
		for name, val := range value {
			switch eventType {
			case "PushEvent":
				outputs = append(outputs, fmt.Sprintf(" - Pushed %d commits to %s", val, name))
			case "PublicEvent":
				outputs = append(outputs, fmt.Sprintf(" - Made %s public, woohoo!", name))
			case "PullRequestEvent":
				outputs = append(outputs, fmt.Sprintf(" - Opened a pull request in %s", name))
			case "CreateEvent":
				outputs = append(outputs, fmt.Sprintf(" - Created %s", name))
			}
		}
	}

	return outputs
}

func getEventsForEachRepo(events []Event) map[string]map[string]int {
	eventsForUser := make(map[string]map[string]int)

	for _, event := range events {
		_, exists := eventsForUser[event.Type]
		if !exists {
			eventsForUser[event.Type] = make(map[string]int)
		}
		_, e := eventsForUser[event.Type][event.Repo.Name]
		if !e {
			eventsForUser[event.Type][event.Repo.Name] = 1
		} else {
			eventsForUser[event.Type][event.Repo.Name] += 1
		}
	}

	return eventsForUser
}

func increment(i int) int {
	i += 1
	return i
}

func callEndpoint(token string, username string) []Event {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	formattedUrl := fmt.Sprintf("https://api.github.com/users/%s/events", username)

	req, err := http.NewRequest("GET", formattedUrl, nil)

	if err != nil {
		log.Fatal(err)
	}

	req.Header.Add("Accept", "application/vnd.github+json")
	req.Header.Add("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Unexpected error has occured, look at the error for further details\n%d", resp.StatusCode)
	}

	var listOfEvents []Event
	err = json.NewDecoder(resp.Body).Decode(&listOfEvents)

	if err != nil {
		log.Fatal(err)
	}

	return listOfEvents
}
