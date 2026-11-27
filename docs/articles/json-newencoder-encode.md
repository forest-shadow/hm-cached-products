# JSON-ответ в Go: `NewEncoder` и `Encode`

**Контекст:** Go 1.27.3, стандартный пакет `encoding/json`, обработчик `GET /health` из `internal/adapters/httpapi/hanler.go`.

## Назначение

`json.NewEncoder(w)` создаёт кодировщик JSON, который будет записывать данные в `w`. Сам вызов `NewEncoder` ещё ничего не записывает. Параметр `w` имеет тип `io.Writer`; в HTTP-обработчике им служит `http.ResponseWriter`.

`Encode(value)` преобразует переданное **Go-значение** в JSON, записывает его в `w`, добавляет в конце перевод строки `\n` и возвращает `error`. Тип параметра — `any`, но не каждое Go-значение можно представить в JSON: например, функцию или канал закодировать нельзя.

## Что происходит в `/health`

```go
type healthResponse struct {
	Status string `json:"status"`
}

w.Header().Set("Content-Type", "application/json; charset=utf-8")
w.WriteHeader(http.StatusOK)
_ = json.NewEncoder(w).Encode(healthResponse{Status: "Ok"})
```

### Короткое объяснение

1. `healthResponse{Status: "Ok"}` создаёт значение структуры и заполняет её поле `Status`.
2. `json.NewEncoder(w)` указывает, куда писать результат: в тело HTTP-ответа.
3. `Encode(...)` превращает структуру в JSON. Тег `json:"status"` задаёт имя поля в JSON — `status` вместо `Status`.
4. В тело ответа попадают байты `{"status":"Ok"}\n`. Перевод строки не меняет значение JSON-документа.

Это поведение следует из правил `encoding/json`; строка `_ =` в текущем обработчике **игнорирует** возможную ошибку `Encode`.

## Какие значения передавать для JSON-объекта

Ниже три обычных способа получить один и тот же объект `{"status":"Ok"}`. Во всех примерах `w` — HTTP-ответ, а после объекта `Encode` добавит `\n`.

### Именованная структура

```go
type healthResponse struct {
	Status string `json:"status"`
}

json.NewEncoder(w).Encode(healthResponse{Status: "Ok"})
```

Это вариант из проекта. Тип заранее описывает форму ответа и удобен, если она используется в нескольких местах. Поле должно быть экспортируемым (`Status` с заглавной буквы), чтобы стандартный кодировщик включил его в JSON. Без тега имя поля в результате было бы `"Status"`.

Значение можно сначала сохранить в переменную и передать её в `Encode`:

```go
response := healthResponse{Status: "Ok"}
json.NewEncoder(w).Encode(response)
```

### Анонимная структура

```go
json.NewEncoder(w).Encode(struct {
	Status string `json:"status"`
}{Status: "Ok"})
```

Тот же принцип, но отдельное имя типа не требуется. Подходит для разового небольшого ответа.

### Map

```go
json.NewEncoder(w).Encode(map[string]string{"status": "Ok"})
```

Здесь `"status"` — строковый ключ map, поэтому тег не нужен. Если значения полей имеют разные типы, можно использовать `map[string]any`, например `map[string]any{"status": "Ok", "ready": true}`. Map удобна, когда набор полей определяется во время выполнения; структура явно задаёт ожидаемые поля и их типы.

## Что не даст такой объект

```go
json.NewEncoder(w).Encode(`{"status":"Ok"}`) // JSON-строка, а не объект
json.NewEncoder(w).Encode("Ok")              // JSON-строка "Ok"
json.NewEncoder(w).Encode([]string{"Ok"})     // JSON-массив ["Ok"]
```

В первом случае готовый текст JSON передан как Go-строка. `Encode` закодирует **строку**, включая кавычки внутри неё; результатом будет `"{\"status\":\"Ok\"}"`. Если JSON уже подготовлен и нужно записать его как есть, можно писать байты непосредственно в `w` или использовать `json.RawMessage`.

Передать `healthResponse` без `{...}` тоже нельзя: это имя **типа**, а `Encode` принимает **значение**.

## HTTP и ошибки

- Заголовок `Content-Type` задаётся отдельно: `Encode` его не устанавливает.
- Для успешного ответа `w.WriteHeader(http.StatusOK)` необязателен: первая запись в `ResponseWriter` установит статус `200 OK` автоматически.
- Ошибку `Encode` следует учитывать: она возможна при кодировании или записи ответа. После начала записи нельзя надёжно заменить уже отправленный ответ другим статусом и телом. В текущем `/health` статус `200 OK` выставляется ещё до вызова `Encode`.

## Главное

`NewEncoder` выбирает место записи; `Encode` принимает Go-значение и создаёт из него JSON. Для объекта `{"status":"Ok"}` подходят структура с тегом, анонимная структура или map с ключом `"status"`. Выбирайте форму Go-значения по тому, насколько явно нужно описать схему ответа и меняется ли набор полей.

**Справка:** [`json.NewEncoder`](https://pkg.go.dev/encoding/json#NewEncoder), [`Encoder.Encode`](https://pkg.go.dev/encoding/json#Encoder.Encode), [`http.ResponseWriter`](https://pkg.go.dev/net/http#ResponseWriter).
