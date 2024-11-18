# Задание L0

## Задание

Необходимо разработать демонстрационный сервис с простейшим интерфейсом, отображающий данные о заказе. Модель данных можно найти в конца задания.

Что нужно сделать:

1. Развернуть локально PostgreSQL
    1. Создать свою БД
    2. Настроить своего пользователя
    3. Создать таблицы для хранения полученных данных
2. Разработать сервис
	1. Реализовать подключение к брокерам и подписку на топик orders в Kafka
	2. Полученные данные записывать в БД
	3. Реализовать кэширование полученных данных в сервисе (сохранять in memory)
	4. В случае прекращения работы сервиса необходимо восстанавливать кэш из БД
	5. Запустить http-сервер и выдавать данные по id из кэша
3. Разработать простейший интерфейс отображения полученных данных по id заказа

## Советы

1. Данные статичны, исходя из этого подумайте насчет модели хранения в кэше и в PostgreSQL. Модель в файле model.json
2. Подумайте как избежать проблем, связанных с тем, что в канал могут закинуть что-угодно
3. Чтобы проверить работает ли подписка онлайн, сделайте себе отдельный скрипт, для публикации данных в топик
4. Подумайте как не терять данные в случае ошибок или проблем с сервисом
5. Apache Kafka разверните локально
6. Удобно разворачивать сервисы для тестирования можно через docker-compose
7. Для просмотра сообщений в Apache Kafka можно пользоваться Redpanda Console

## Бонус-задание

1. Покройте сервис автотестами — будет плюсик вам в карму
2. Устройте вашему сервису стресс-тест: выясните на что он способен
3. Логи в JSON-формате делают их более удобными для машинной обработки 

Воспользуйтесь утилитами WRK и Vegeta, попробуйте оптимизировать код.

## Результат

По готовности сервиса снимите короткое видео работы интерфейса и вместе со ссылкой на репозиторий пришлите на проверку через личный кабинет.

## Модель данных

```go
{

   "order_uid": "b563feb7b2b84b6test",
   "track_number": "WBILMTESTTRACK",
   "entry": "WBIL",
   "delivery": {
      "name": "Test Testov",
      "phone": "+9720000000",
      "zip": "2639809",
      "city": "Kiryat Mozkin",
      "address": "Ploshad Mira 15",
      "region": "Kraiot",
      "email": "test@gmail.com"
   },
   "payment": {
      "transaction": "b563feb7b2b84b6test",
      "request_id": "",
      "currency": "USD",
      "provider": "wbpay",
      "amount": 1817,
      "payment_dt": 1637907727,
      "bank": "alpha",
      "delivery_cost": 1500,
      "goods_total": 317,
      "custom_fee": 0
   },
   "items": [
      {
         "chrt_id": 9934930,
         "track_number": "WBILMTESTTRACK",
         "price": 453,
         "rid": "ab4219087a764ae0btest",
         "name": "Mascaras",
         "sale": 30,
         "size": "0",
         "total_price": 317,
         "nm_id": 2389212,
         "brand": "Vivienne Sabo",
         "status": 202
      }
   ],
   "locale": "en",
   "internal_signature": "",
   "customer_id": "test",
   "delivery_service": "meest",
   "shardkey": "9",
   "sm_id": 99,
   "date_created": "2021-11-26T06:22:19Z",
   "oof_shard": "1"
}
```