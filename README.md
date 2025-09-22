# API Документация

## Создание подписки

```bash
curl -X POST http://localhost:8080/subscriptions \
  -H "Content-Type: application/json" \
  -d '{
    "service_name": "Yandex Plus",
    "price": 400,
    "user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
    "start_date": "07-2025"
  }'
```

## Получение списка подписок

```bash
curl http://localhost:8080/subscriptions
```

## Получение подписки по ID

```bash
curl http://localhost:8080/subscriptions/{id}
```

## Обновление подписки

```bash
curl -X PUT http://localhost:8080/subscriptions/{id} \
  -H "Content-Type: application/json" \
  -d '{
    "service_name": "Yandex Plus Updated",
    "price": 450
  }'
```

## Удаление подписки

```bash
curl -X DELETE http://localhost:8080/subscriptions/{id}
```

## Агрегация стоимости подписок

```bash
curl -X POST http://localhost:8080/subscriptions/aggregate \
  -H "Content-Type: application/json" \
  -d '{
    "start_date": "01-2025",
    "end_date": "12-2025",
    "user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba"
  }'
```
### Фильтры для агрегации

- `user_id` (optional) - фильтрация по пользователю
- `service_name` (optional) - фильтрация по названию сервиса
- `start_date` (required) - начало периода (MM-YYYY)
- `end_date` (required) - конец периода (MM-YYYY)



