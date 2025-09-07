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
- PostgreSQL — СУБД
- Docker + Docker Compose
- Swagger/OpenAPI — документация API
- slog — структурированный логгер

# Запуск проекта
docker compose up --build

Swagger UI: http://localhost:8080/docs

# Примеры запросов

1) создать подписку:

curl -s -X POST http://localhost:8080/subscriptions \
-H "Content-Type: application/json" \
-d '{
"user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
"service_name": "Netflix",
"price": 499,
"start": "07-2025"
}' | jq

2) получить список подписок:

curl -s "http://localhost:8080/subscriptions?limit=10&offset=0" | jq

3) получить подписку по id: 

curl -s http://localhost:8080/subscriptions/$ID | jq

4) обновить подписку 

curl -s -X PUT http://localhost:8080/subscriptions/$ID \
-H "Content-Type: application/json" \
-d '{
"user_id":"'"$USER_ID"'",
"service_name":"Test Update Subscription",
"price":699,
"start":"08-2025"
}' | jq

5) удалить подписку

curl -i -X DELETE http://localhost:8080/subscriptions/$ID

6) получение списка подписок (с фильтрами):

6.1) получить все подписки:

   curl -s "http://localhost:8080/subscriptions?limit=10&offset=0" | jq

6.2) фильтр по пользователю user_id:

   curl -s "http://localhost:8080/subscriptions?user_id=$USER_ID&limit=10&offset=0" | jq
   
6.3) фильтр по названию сервиса service_name:

   curl -s "http://localhost:8080/subscriptions?service_name=Netflix&limit=10&offset=0" | jq

7) сумма стоимости подписок (с фильтрами):

7.1) сумма стоимости всех подписок:

   curl -s "http://localhost:8080/subscriptions/summary?from=06-2025&to=08-2025" | jq

7.2) сумма стоимости подписок пользователя:

curl -s "http://localhost:8080/subscriptions/summary?from=06-2025&to=08-2025&user_id=$USER_ID" | jq

7.3) сумма стоимости подписок по сервису:

curl -s "http://localhost:8080/subscriptions/summary?from=06-2025&to=08-2025&service_name=Netflix" | jq