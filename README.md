# LibraTrack — Тема 9: автентифікація, JWT та RBAC

REST API бібліотеки (Go, Gin, GORM, PostgreSQL), захищений за допомогою bcrypt, JWT і ролей.

## Запуск

```bash
docker compose up -d                      # PostgreSQL на порту 5434
cp .env.example .env                      # задайте JWT_SECRET (мінімум 32 символи)
export $(grep -v '^#' .env | xargs)       # або задайте змінні вручну
go run ./cmd/api                          # міграції застосовуються при старті
```

Swagger UI: <http://localhost:8080/swagger>.

## Тести

```bash
export TEST_DATABASE_URL='postgres://libratrack:libratrack@localhost:5434/libratrack_test?sslmode=disable'
go test -race ./...
```

Без `TEST_DATABASE_URL` тести з БД пропускаються.

## Безпека

| Документ | Зміст |
|---|---|
| [security_design.txt](security_design.txt) | ролі, матриця доступу, DTO |
| [prompts.txt](prompts.txt) | промпти завдань 2.1 і 2.2 |
| [security_audit.md](security_audit.md) | аудит: знахідки, серйозність, тріаж |
| [Homework_tasks_9.md](Homework_tasks_9.md) | текст завдання |

Ролі: `reader`, `librarian`, `admin`. Нова реєстрація завжди отримує `reader`.

| Група | Маршрути | Доступ |
|---|---|---|
| public | `POST /auth/register`, `POST /auth/login`, `GET` каталогу | без токена |
| protected | `GET/PUT /me`, `GET /me/loans` | будь-який користувач з токеном |
| staff | запис каталогу, `POST /loans`, `POST /loans/:id/return` | `librarian`, `admin` |
| admin | `GET /users`, `PUT /users/:id/role` | `admin` |

Без токена — `401`, роль не підходить — `403`.
