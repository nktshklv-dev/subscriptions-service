REST API сервис для управления пользовательскими подписками

# Возможности

- Создание подписки
- Получение подписки по ID
- Обновление подписки
- Удаление подписки
- Список подписок с фильтрацией и пагинацией
- Подсчёт суммарной стоимости подписок за период

# Технологии

- Go — HTTP-сервер
- PostgreSQL — база данных
- Docker + Docker Compose
- Swagger/OpenAPI — документация API
- slog — структурированный логгер

# Запуск проекта
docker compose up --build

Swagger UI: http://localhost:8080/docs

Примеры запросов
1) создать подписку 
curl -s -X POST http://localhost:8080/subscriptions \
-H "Content-Type: application/json" \
-d '{
"user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
"service_name": "Netflix",
"price": 499,
"start": "07-2025"
}' | jq

2) получить список подписок
   curl -s "http://localhost:8080/subscriptions?limit=10&offset=0" | jq