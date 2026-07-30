# auction-house-lotTrio

[![Lint](https://github.com/msdilnozabonu/auction-house-lotTrio/actions/workflows/lint.yml/badge.svg?branch=develop)](https://github.com/msdilnozabonu/auction-house-lotTrio/actions/workflows/lint.yml)
[![Nilaway](https://github.com/msdilnozabonu/auction-house-lotTrio/actions/workflows/nilaway.yml/badge.svg?branch=develop)](https://github.com/msdilnozabonu/auction-house-lotTrio/actions/workflows/nilaway.yml)
[![Coverage](https://github.com/msdilnozabonu/auction-house-lotTrio/blob/badges/coverage.svg)](https://github.com/msdilnozabonu/auction-house-lotTrio/blob/badges/coverage.svg)
[![Go Version](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go)](https://go.dev/)

## Описание
**AuctionHouse** — backend-платформа онлайн-аукциона на Go, построенная на основе Layered Architecture.
Включает JWT-аутентификацию, ролевую модель доступа, управление лотами, систему ставок, модерацию, аналитику, автоматическое
закрытие аукционов, экспорт отчётов, PostgreSQL, Docker и REST API.

---
## Содержание

- [Ключевые особенности](#ключевые-особенности)
- [Архитектура](#архитектура)
- [Установка и запуск](#-установка-и-запуск)
- [Переменные окружения](#-переменные-окружения)
- [Миграции](#миграции)
- [Swagger](#-swagger)
- [Роли и возможности](#-роли-и-возможности)
- [Проверка проекта](#-проверка-проекта)
- [Полезные команды](#полезные-команды)
- [Стек технологий](#стек-технологий)
- [Команда](#команда)
---


## Ключевые особенности

- **JWT-аутентификация (Access / Refresh Token)** - Безопасная аутентификация пользователей
- **Регистрация и авторизация пользователей**
- **Ролевая модель доступа** - Управление правами доступа на основе ролей
- **Управление лотами** - Создание, редактирование и управление аукционными лотами
- **Система ставок** - Полнофункциональная система ставок с валидацией
- **Автоматическое закрытие аукционов** - Автоматизированное управление сроками
- **Аналитика** — агрегаты по продажам продавца и по площадке в целом
- **Загрузка фото лота** (multipart), watchlist, история ставок, экспорт отчётов
- **Единый формат ошибок и ответов**, structured logging (`slog`), graceful shutdown
- **Swagger API**
- **Docker**
- **Middleware (логирование, RequestID, авторизация)**


## Архитектура

Однонаправленный поток зависимостей — только вниз, через интерфейсы:

```
handler  →  service  →  repository  →  PostgreSQL
```


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

### Локальный запуск

```bash
go run ./cmd/server
```

Поднимаются `app` + `postgres`. Проверка: `GET /health` → 200.

---

## 🔧 Переменные окружения

Для настройки проекта используется файл `.env`.

Пример:

```env
PORT=9999

DB_DSN=postgres://app:pass@localhost:5445/auction_house

JWT_SECRET=super-secret-key
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

## 👥 Роли и возможности

| Роль | Ключевые эндпоинты |
|------|---------------------|
| 🙋 **Участник (bidder)** | каталог лотов с фильтрами/поиском, ставка (`POST /lots/:id/bid`), мои ставки и выигрыши, watchlist, история ставок по лоту |
| 📦 **Продавец (seller)** | CRUD лотов, статусы `draft/live/closed`, загрузка фото, свои лоты и ставки по ним, аналитика продаж, отмена лота без ставок |
| ⚙️ **Платформа (admin)** | панель всех лотов, модерация (approve/reject), авто- и ручное закрытие лотов (`POST /admin/close-expired`), аналитика площадки, экспорт отчётов, разбор жалоб |

---

## 🧪 Проверка проекта

Линтер

```bash
golangci-lint run
```

Тесты

```bash
go test -tags=integration ./...
```

Покрытие

```bash
go test -tags=integration ./... -cover
```

Race

```bash
go test -race ./...
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


## Команда

Проект разработан командой из трёх участников.
Ответственность была разделена по функциональным областям:

- Участник (Bidder)
- Продавец (Seller)
- Платформа (Admin)