# pr-reviewer

Сервис назначения ревьюверов для pull request’ов внутри команд.

* хранит команды, пользователей и PR в PostgreSQL
* при создании PR автоматически выбирает до двух активных ревьюверов из команды автора (без него самого)
* поддерживает merge (идемпотентный) и безопасный reassign ревьювера
* отдаёт метрики в формате Prometheus


## Cтарт

```bash
docker compose up --build
```

После этого:
* API доступен на `http://localhost:8080`
* метрики — на `http://localhost:8080/metrics`

Остановка:

```bash
docker compose down
```

## Тесты

### Юнит- и HTTP-тесты

```bash
go test ./...
```

(используются мок-объекты из `mocks/` и `testify`)

### E2E-тесты (живой сервис + Postgres)

```bash
cd e2e
docker compose -f docker-compose.e2e.yaml up --build --abort-on-container-exit
docker compose -f docker-compose.e2e.yaml down
```

E2E-тесты поднимают отдельную БД, сервис и прогоняют сценарии через HTTP.


## Метрики и observability

* эндпоинт метрик: `GET /metrics`
* основные метрики:

  * `http_requests_total{method,path,status}`
  * `http_request_duration_seconds{method,path,status}`
  * `pr_events_total{event="created|merged"}`
  * `pr_reassign_total{result="success|not_found|pr_merged|not_assigned|no_candidate|internal_error"}`

Можно скрапить Prometheus’ом, добавив таргет `http://localhost:8080/metrics`.

## Эндпоинты

* `POST /team/add` — создать/обновить команду
* `GET  /team/get` — получить команду и участников
* `POST /users/setIsActive` — активировать/деактивировать пользователя
* `GET  /users/getReview` — PR, где пользователь выступает ревьювером
* `POST /pullRequest/create` — создать PR и автоматически назначить ревьюверов
* `POST /pullRequest/merge` — смерджить PR (идемпотентно)
* `POST /pullRequest/reassign` — переназначить одного ревьювера

