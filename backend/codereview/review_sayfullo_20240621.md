## Code review Sayfullokhon

### Create chracter
#### `@app.post("/api/characters")`
Тело сообщение должно содержать следующие поля:

```json
{
    "name": "Gladiator's Name"
}
```
Затем нужно создать персонажа в базе данных с базовыми значениями.

Ответ должен быть в формате JSON, и содержать следующие поля: 
```json
{
    "id": 1,
    "name": "Gladiator's Name",
    "strength": 0,
    "agility": 0,
    "stamina": 0,
    "level": 1,
    "availablePoints": 5
}
```

### Get chracter
#### @app.get("/api/characters/{id}")
Если персонажа нет в базе данных, нужно вернуть ошибку 404

```py
    return {"response": "Character not found"}
```
Данный код вернет статус 200 и это значит что персонаж найден

Нужно вернуть исключение со статусом 404

```py
    raise HTTPException(status_code=404, detail="Character not found")
```

### Update chracter
#### @app.put("/api/characters/{id}")
Если персонажа нет в базе данных, нужно вернуть ошибку 404 как и с GET

```py
    raise HTTPException(status_code=404, detail="Character not found")
```
