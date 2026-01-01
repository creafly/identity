# Identity Service

Сервис управления пользователями, тенантами, ролями и правами доступа.

## Структура проекта

```
identity/
├── cmd/
│   └── api/              # Точка входа приложения
├── internal/
│   ├── config/           # Конфигурация
│   ├── domain/
│   │   ├── entity/       # Доменные модели
│   │   ├── repository/   # Интерфейсы репозиториев
│   │   └── service/      # Бизнес-логика
│   ├── handler/          # HTTP обработчики
│   ├── middleware/       # Middleware (auth, locale)
│   └── i18n/             # Локализация
├── migrations/           # SQL миграции
├── pkg/
│   └── utils/            # Вспомогательные функции
└── resources/            # Файлы локализации
```

## Технологии

- **Go 1.22+** - язык программирования
- **Gin** - HTTP фреймворк
- **PostgreSQL** - база данных
- **sqlx** - работа с SQL
- **JWT** - аутентификация

## API Endpoints

### Публичные

- `POST /api/v1/auth/register` - Регистрация
- `POST /api/v1/auth/login` - Вход
- `POST /api/v1/auth/refresh` - Обновление токена

### Защищённые (требуют Authorization: Bearer <token>)

- `GET /api/v1/me` - Текущий пользователь
- `POST /api/v1/change-password` - Смена пароля

### Health checks

- `GET /health` - Статус сервиса
- `GET /ready` - Готовность к работе

## Локализация

Сервис поддерживает локализацию через заголовок `Accept-Language` или query параметр `locale`.

Поддерживаемые языки:
- `en-US` - English (по умолчанию)
- `ru-RU` - Русский

## Установка

```bash
# Установка зависимостей
make deps

# Сборка
make build

# Запуск
make run
```

## Конфигурация

Создайте `.env` файл:

```env
PORT=8080
HOST=0.0.0.0
GIN_MODE=debug

DATABASE_URL=postgres://postgres:postgres@localhost:5432/identity?sslmode=disable

JWT_SECRET=your-super-secret-key
JWT_ACCESS_TOKEN_DURATION=15m
JWT_REFRESH_TOKEN_DURATION=168h

DEFAULT_LOCALE=en-US
LOG_LEVEL=debug
```

## Миграции

```bash
# Применить миграции
make migrate
```

## Docker

```bash
docker build -t identity .
docker run -p 8080:8080 --env-file .env identity
```
