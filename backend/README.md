## Backend APIs
Здесь приведены возможные АПИ для введения боя гладиаторов

### Create Character

- Endpoint: `/api/characters`
- Method: `POST`
- Request Body: `{ "name": "Gladiator's Name" }`
- Response: `{ "id": 1, "name": "Gladiator's Name", "strength": 0, "agility": 0, "stamina": 0, "level": 1, "availablePoints": 5 }`


### Adjust Character Attributes

- Endpoint: `/api/characters/{id}/attributes`
- Method: `PUT`
- Request Body: `{ "strength": 3, "agility": 1, "stamina": 1 }`
- Response: `{ "id": 1, "name": "Gladiator's Name", "strength": 3, "agility": 1, "stamina": 1, "level": 1, "availablePoints": 0 }`


### Create Lobby

- Endpoint: `/api/lobbies`
- Method: `POST`
- Response: `{ "lobbyId": 1 }`


### Join Lobby

- Endpoint: `/api/lobbies/{lobbyId}/join`
- Method: `POST`
- Request Body: `{ "characterId": 1 }`
- Response: `{ "lobbyId": 1, "players": [{ "id": 1, "name": "Gladiator's Name", "strength": 3, "agility": 1, "stamina": 1, "level": 1, "availablePoints": 0 }] }`


### Start Fight

- Endpoint: `/api/lobbies/{lobbyId}/fights`
- Method: `POST`
- Response: `{ "fightId": 1, "playerTurn": true }`


### Make Move

- Endpoint: `/api/fights/{fightId}/moves`
- Method: `POST`
- Request Body: `{ "attackArea": "head", "blockAreas": ["chest", "groin"] }`
- Response: `{ "playerHit": true, "opponentHit": false, "playerDamageDealt": 10, "opponentDamageDealt": 0, "playerHealth": 95, "opponentHealth": 80 }`


### End Fight

- Endpoint: `/api/fights/{fightId}`
- Method: `DELETE`
- Response: `{ "winner": "player", "experience": 100 }`
