# auction-house-lotTrio

[![Lint](https://github.com/msdilnozabonu/auction-house-lotTrio/actions/workflows/lint.yml/badge.svg?branch=develop)](https://github.com/msdilnozabonu/auction-house-lotTrio/actions/workflows/lint.yml)
[![Nilaway](https://github.com/msdilnozabonu/auction-house-lotTrio/actions/workflows/nilaway.yml/badge.svg?branch=develop)](https://github.com/msdilnozabonu/auction-house-lotTrio/actions/workflows/nilaway.yml)
[![Coverage](https://github.com/msdilnozabonu/auction-house-lotTrio/blob/badges/coverage.svg)](https://github.com/msdilnozabonu/auction-house-lotTrio/blob/badges/coverage.svg)

## Описание
**AuctionHouse** — backend-платформа онлайн-аукциона на Go, построенная на основе Layered Architecture. 
Включает JWT-аутентификацию, ролевую модель доступа, управление лотами, систему ставок, автоматическое 
закрытие аукционов, PostgreSQL, Docker и REST API.


## Ключевые особенности

- **JWT-аутентификация (Access / Refresh Token)** - Безопасная аутентификация пользователей
- **Регистрация и авторизация пользователей**
- **Ролевая модель доступа** - Управление правами доступа на основе ролей
- **Управление лотами** - Создание, редактирование и управление аукционными лотами
- **Система ставок** - Полнофункциональная система ставок с валидацией
- **Автоматическое закрытие аукционов** - Автоматизированное управление сроками
- **Swagger API**
- **Docker**
- **Middleware (логирование, RequestID, авторизация)**

## 📦 Установка и запуск

### Клонирование репозитория

```bash
git clone https://github.com/msdilnozabonu/auction-house-lotTrio.git
cd auction-house-lotTrio
```

### Запуск через Docker

```bash
docker compose up --build
```

или локально

```bash
go run ./cmd/server
```

---

## 🔧 Переменные окружения

Для настройки проекта используется файл `.env`.

Пример:

```env
PORT=9999

DB_DSN=postgres://app:pass@localhost:5445/auction_house

JWT_SECRET=your_secret_key
```

---

## Миграции

Для применения миграций:

```bash
migrate -path migrations/server -database "postgres://app:pass@localhost:5445/auction_house?sslmode=disable" up
```

---

## 📖 Swagger

После запуска проекта документация доступна по адресу:

```
http://localhost:9999/api/v1/swagger/index.html
```

---

## 🧪 Проверка проекта

Линтер

```bash
golangci-lint run
```

Тесты

```bash
go test ./...
```

Покрытие

```bash
go test ./... -cover
```

Nilaway

```bash
nilaway ./...
```
---

## Полезные команды

```bash
make build     # собрать проект
make run       # запустить сервер
make fmt       # форматирование кода
make test      # запустить тесты
make cover     # проверить покрытие
make lint      # запустить линтеры
make check     # выполнить все проверки
```

## Стек технологий

- **Go** - Язык программирования
- **Gin** - Веб-фреймворк
- **PostgreSQL** - База данных
- **Docker** - Контейнеризация
- **Swagger** - Документация API
- **GitHub Actions** - CI/CD
- **Layered Architecture** - Архитектурный паттерн
