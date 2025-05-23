package main

import (
	"fmt"
	"gladiator-backend/api"
	"log"
	"net/http"
	"strings"
)

func main() {
	// Character routes
	http.HandleFunc("/api/characters", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			api.CreateCharacter(w, r)
		} else {
			http.Error(w, "Use POST to create a character at /api/characters", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/api/characters/", func(w http.ResponseWriter, r *http.Request) {
		pathParts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
		// Expected: /api/characters/{id}
		// Expected: /api/characters/{id}/attributes

		if len(pathParts) < 4 || pathParts[3] == "" { // {id} part
			http.Error(w, "Character ID is missing in URL path", http.StatusBadRequest)
			return
		}

		if len(pathParts) == 4 { // Path is /api/characters/{id}
			if r.Method == http.MethodGet {
				api.GetCharacter(w, r)
			} else {
				http.Error(w, "Use GET to retrieve a character at /api/characters/{id}", http.StatusMethodNotAllowed)
			}
		} else if len(pathParts) == 5 && pathParts[4] == "attributes" { // Path is /api/characters/{id}/attributes
			if r.Method == http.MethodPut {
				api.UpdateCharacterAttributes(w, r)
			} else {
				http.Error(w, "Use PUT to update attributes at /api/characters/{id}/attributes", http.StatusMethodNotAllowed)
			}
		} else {
			http.NotFound(w, r)
		}
	})

	// Lobby routes
	http.HandleFunc("/api/lobbies", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			api.CreateLobby(w, r)
		} else {
			http.Error(w, "Use POST to create a lobby at /api/lobbies", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/api/lobbies/", func(w http.ResponseWriter, r *http.Request) {
		pathParts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
		// Expected:
		// /api/lobbies/{lobbyId}/join (POST)
		// /api/lobbies/{lobbyId}/fights (POST)

		if len(pathParts) < 5 || pathParts[3] == "" { // {lobbyId} part
			http.Error(w, "Lobby ID is missing or path is malformed", http.StatusBadRequest)
			return
		}

		action := pathParts[4] // "join" or "fights"

		if action == "join" {
			if r.Method == http.MethodPost {
				api.JoinLobby(w, r)
			} else {
				http.Error(w, "Use POST to join a lobby at /api/lobbies/{lobbyId}/join", http.StatusMethodNotAllowed)
			}
		} else if action == "fights" {
			if r.Method == http.MethodPost {
				api.StartFight(w, r)
			} else {
				http.Error(w, "Use POST to start a fight at /api/lobbies/{lobbyId}/fights", http.StatusMethodNotAllowed)
			}
		} else {
			http.NotFound(w, r)
		}
	})

	// Fight routes
	http.HandleFunc("/api/fights/start-ai", api.StartAIFight) // New route for starting AI fight

	http.HandleFunc("/api/fights/", func(w http.ResponseWriter, r *http.Request) {
		pathParts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
		// Expected:
		// /api/fights/{fightId}/moves (POST)
		// /api/fights/{fightId} (DELETE for EndFight)

		if len(pathParts) < 4 || pathParts[3] == "" { // {fightId} part
			http.Error(w, "Fight ID is missing in URL path", http.StatusBadRequest)
			return
		}
		// fightID := pathParts[3] // Handled by the specific handlers

		if len(pathParts) == 4 { // Path is /api/fights/{fightId}
			if r.Method == http.MethodDelete {
				api.EndFight(w, r)
			} else {
				http.Error(w, "Use DELETE to end a fight at /api/fights/{fightId}", http.StatusMethodNotAllowed)
			}
		} else if len(pathParts) == 5 && pathParts[4] == "moves" { // Path is /api/fights/{fightId}/moves
			if r.Method == http.MethodPost {
				api.MakeMove(w, r)
			} else {
				http.Error(w, "Use POST to make a move at /api/fights/{fightId}/moves", http.StatusMethodNotAllowed)
			}
		} else {
			http.NotFound(w, r)
		}
	})

	port := "8080"
	fmt.Printf("Server starting on port %s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start server: %s\n", err)
	}
}
