# WarehouseControl

Мини-система учёта товаров на складе с CRUD-операциями, ролевой моделью доступа,
JWT-авторизацией, аудитом изменений через триггеры PostgreSQL и веб-интерфейсом.


> **Учебный проект.** Аудит намеренно реализован через триггеры PostgreSQL — это антипаттерн,
> демонстрирующий, почему бизнес-логику не следует выносить в СУБД.
> 

## Возможности

- **CRUD товаров** — создание, просмотр, редактирование (partial update), удаление
- **JWT-авторизация** — роль зашивается в токен, проверяется на каждом запросе
- **Ролевая модель** — admin, manager, viewer с разграничением прав
- **Аудит изменений** — автоматическое логирование INSERT/UPDATE/DELETE через триггер PostgreSQL
- **Diff между версиями** — для каждого UPDATE сохраняется JSON-diff изменённых полей
- **Фильтрация аудита** — по дате, пользователю, действию, товару
- **Экспорт в CSV** — выгрузка истории изменений
- **Поиск товаров** — по названию и SKU
- **Пагинация** — для списков товаров и аудита
- **Веб-интерфейс** — просмотр, редактирование товаров, история изменений
- **Graceful shutdown** — корректное завершение по SIGINT/SIGTERM
- **Retry-стратегия** — повторные попытки при сбоях БД
- **Structured logging** — структурированные логи с request ID

---

## Технологии

| Компонент       | Технология               |
|-----------------|--------------------------|
| Язык            | Go 1.25                  |
| HTTP-фреймворк  | Gin (ginext wbf)         |
| База данных     | PostgreSQL 17 (dbpg wbf) |
| Миграции        | Goose                    |
| Авторизация     | JWT (golang-jwt/jwt)     |
| Хеширование     | bcrypt                   |
| Контейнеризация | Docker, Docker Compose   |
| Логирование     | slog (Logger wbf)        |

## Быстрый старт


### Запуск через Docker Compose (рекомендуется)

```bash
# Клонировать репозиторий
git clone https://github.com/stpnv0/WarehouseControl.git
cd WarehouseControl

# Создать конфигурационные файлы
cp .env.example .env
cp config/config.yaml.example config/config.yaml

# Поднять всё
docker compose up -d --build
```
Приложение будет доступно на http://localhost:8080


### Тестовые пользователи
| Логин   | Пароль    | Роль    |
|---------|-----------|---------|
| admin   | password  | admin   |
| manager | password  | manager |
| viewer  | password  | viewer  |


## Аудит через триггеры

Аудит реализован через PostgreSQL-триггер fn_item_audit(), который срабатывает на AFTER INSERT OR UPDATE OR DELETE таблицы items
### Как это работает
1) Перед выполнением CUD-операции приложение открывает транзакцию
2) Устанавливает сессионную переменную:
   ```
    SELECT set_config('app.current_user_id', '<uuid>', true);
   ```
3) Выполняет INSERT/UPDATE/DELETE
4) Триггер автоматически пишет запись в item_audit_log с:
   - old_data — состояние до изменения (JSONB)
   - new_data — состояние после изменения (JSONB)
   - diff — только изменившиеся поля (JSONB)
   - changed_by — UUID пользователя из сессионной переменной

## Структура проекта

```
WarehouseControl/
├── cmd/
│   └── warehouse/                  # точка входа
├── config/                         # конфигурация
├── internal/
│   ├── app/                        # инициализация, DI, запуск, shutdown
│   ├── auth/                       # JWT: генерация и валидация токенов
│   ├── config/                     # структуры конфигурации, загрузка
│   ├── domain/                     # доменные модели и ошибки
│   ├── handler/                    # HTTP-обработчики и DTO
│   ├── export/                     # экспорт аудита в CSV
│   ├── middleware/                 # JWT, CORS, логгирование, X-Request-ID
│   ├── repository/                 # доступ к БД
│   ├── router/                     # маршрутизация, middleware
│   └── service/                    # бизнес-логика
├── migrations/                     # SQL-миграции (goose)
├── web/                            # фронтенд
├── .env.example
├── .gitignore
├── docker-compose.yml
├── Dockerfile
├── go.mod
├── go.sum
└── README.md
```