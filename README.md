# WASAText

A full-stack messaging web application: one-to-one and group chats, photo messages, replies, forwarding, emoji reactions and read receipts.

Built with a **Go** REST API, a **SQLite** database and a **Vue 3** single-page frontend, packaged as two **Docker** images and started together with **Docker Compose**. Developed as the project for the *Web and Software Architecture* course at Sapienza University of Rome.

## Features

- Username-based login (users are created on first login)
- Private conversations and groups, sorted by latest activity
- Text and photo messages, replies, forwarding, deletion
- Emoji reactions
- Read receipts: one checkmark when every recipient received the message, two when every recipient read it
- Groups: create, add members, leave, rename, change photo, member list
- User search, editable username and profile photo
- Near real-time updates via polling

## Tech stack

| Layer | Technologies |
|---|---|
| Backend | Go, httprouter, SQLite (go-sqlite3) |
| API | REST, documented with OpenAPI 3 (`doc/api.yaml`) |
| Frontend | Vue 3, Vue Router, axios, Bootstrap 5, Vite |
| Deployment | Docker multi-stage builds, Docker Compose, nginx |

## Design notes

- **API-first:** the OpenAPI specification was written before the implementation and every endpoint follows it.
- **Read receipts:** a `message_status` table stores one row per recipient (sent / received / read). The status shown to the sender is the minimum across recipients, so in a group the double checkmark appears only when everyone has read the message. Statuses advance as a side effect of fetching the conversation list (received) and opening a conversation (read).
- **Photos:** stored as BLOBs in a `photos` table; JSON responses only carry the photo URL (`/photos/{id}`), which the browser loads directly.
- **Layered backend:** HTTP handlers (`service/api`) are separated from data access (`service/database`), which exposes an `AppDatabase` interface.
- **Small images:** both Dockerfiles are multi-stage, so the final images contain only the compiled binary (on Debian) or the static build (on nginx).

## Quick start (Docker Compose)

```sh
docker compose up --build
```

Open http://localhost:8081. Stop with `Ctrl+C`, then `docker compose down`.

Without Compose, build and run the two images separately:

```sh
docker build -t wasatext-backend -f Dockerfile.backend .
docker build -t wasatext-frontend -f Dockerfile.frontend .

docker run --rm -p 3000:3000 wasatext-backend
docker run --rm -p 8081:80 wasatext-frontend
```

## Development

```sh
# backend (port 3000)
go run ./cmd/webapi/

# frontend (port 5173)
./open-node.sh
yarn run dev
```

## Project structure

```
cmd/webapi/         backend entry point
service/api/        HTTP handlers and routes
service/database/   SQLite schema and queries
doc/api.yaml        OpenAPI specification
webui/              Vue frontend
compose.yaml        runs backend and frontend together
```

## Credits

Based on the [Fantastic Coffee (decaffeinated)](https://github.com/sapienzaapps/fantastic-coffee-decaffeinated) project template by Sapienza University of Rome.
