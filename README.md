# Go Motivation Bot

Telegram-бот для ежедневной рассылки мотивационных цитат. Event-driven архитектура на Go с MongoDB.

## Возможности

- Онбординг через inline-клавиатуры: выбор таймзоны, кол-ва цитат (1-3), дней недели, времени отправки
- Cron-планировщик каждые 5 минут проверяет, кому пора отправить цитату
- Итеративная выдача цитат — у каждого пользователя свой указатель, цитаты идут по порядку
- Админ-панель для добавления цитат через бота (`/add`)
- Event-driven: Cron → TickEvent → SchedulerHandler → QuoteSendRequested → DeliveryHandler

## Быстрый старт

### 1. Настроить конфигурацию

```bash
cp .env.example .env
```

Заполнить `.env`:
- `BOT_TOKEN` — токен от @BotFather
- `MONGO_URI` — URI подключения к MongoDB (по умолчанию `mongodb://mongo:27017`)
- `MONGO_DB` — имя базы данных (по умолчанию `go_motivation`)
- `ADMIN_CHAT_ID` — Telegram chat ID администратора

### 2. Запуск через Docker

```bash
make docker-up
```

Это поднимет MongoDB и бота в Docker-контейнерах.

### 3. Локальный запуск (для разработки)

Запустить MongoDB отдельно, затем:

```bash
make run
```

## Команды бота

### Для всех пользователей
- `/start` — начать настройку / перенастроить расписание
- `/settings` — показать текущие настройки

### Для администратора
- `/add` — добавить цитату (пошаговый ввод) Текст → автор (или Skip) → год (или Skip) → примечания (или Skip) → категория (или Skip) → сохранение.
- `/quotes` — показать количество цитат в базе

## Структура проекта

```
cmd/bot/main.go              — Точка входа, DI (dig)
internal/
  config/                    — Конфигурация (cleanenv)
  entity/                    — Доменные сущности (Quote, User)
  repository/                — Интерфейсы + MongoDB реализация
  service/                   — Бизнес-логика
  event/                     — Event Bus
  handler/bot/               — Обработчики Telegram (роутер, онбординг, админ)
  handler/event/             — Обработчики событий (планировщик, доставка)
  telegram/                  — Обёртка над telegram-bot-api + клавиатуры
  scheduler/                 — Cron-планировщик
```

## Стек

- Go 1.23
- telegram-bot-api/v5
- MongoDB (mongo-driver/v2)
- cleanenv, robfig/cron, logrus, dig
