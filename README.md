# user-balance-service

Микросервис для работы с балансом пользователей (зачисление средств, списание средств, перевод средств от пользователя к пользователю, а также метод получения баланса пользователя). Сервис предоставляет HTTP API и принимает/отдаёт запросы/ответы в формате JSON.

## Операции

#### CREATE
> [auth/sign-up] --  Регистрация юзера (принимает username и password)[POST-запрос] 

> [auth/sign-in] --  Авторизация юзера (хэдер basic auth(username, password) -> jwt-token) [POST-запрос]

> [api/account/create] --  Создание аккаунта с балансом (баланс будет 0) [POST-запрос] 
#### READ:
> [api/account/state] -- Проверка баланса на конкретном аккаунте (принимает id и balance) [GET-запрос]

> [api/account/history/all] -- Просмотр истории операций всех аккаунтов [GET-запрос]

> [api/account/history/:id] -- Просмотр истории операций конкретного аккаунта [GET-запрос]
#### UPDATE:
> [api/account/refill] -- Пополнение баланса аккаунта (принимает id и balance) [PUT-запрос]

> [api/account/write-off] -- Списание с баланса аккаунта (принимает id и balance) [PUT-запрос]

> [api/account/transfer] -- Перевод суммы с одного баланса на другой (принимает id_from, id_to, amount) [PUT-запрос]
#### DELETE:
> [api/account/delete] -- Удаление аккаунта (принимает id) [DELETE-запрос] 

# Примечание: для доступа к эндпоинту /api необходимо ввести в поле bearer token jwt-token, сгенерированный при авторизации 

## Дополнительные возможности:
> [api/account/state?currency=] -- Конвертация баланса аккаунта с рубля на указанную валюту [GET-запрос]

> [api/account/history/all?sort=] -- Сортировка истории операций всех аккаунтов по сумме и дате (принимает amount / date в формате "2022-10-20") [GET-запрос]

> [api/account/history/:id?sort=] -- Сортировка истории операций конкретного аккаунта по сумме и дате (принимает amount / date в формате "2022-10-20") [GET-запрос]

> [api/account/history/:id?limit=&cursor=] -- Вывод истории операций конкретного аккаунта постранично (cursor в формате "2022-10-20") [GET-запрос]

> [api/account/history/all?limit=&cursor=] -- Вывод истории операций всех аккаунтов постранично (cursor в формате "2022-10-20") [GET-запрос]


## Примеры использования:
#### Curl:
> curl --location --request GET 'localhost:8080/api/account/state' \
--header 'Authorization: Bearer {some_token}' \
--header 'Content-Type: application/json' \
--data-raw '{
    "id": 3
}'

> curl --location --request PUT 'localhost:8080/api/account/refill' \
--header 'Authorization: Bearer {some_token}' \
--header 'Content-Type: application/json' \
--data-raw '{
    "id": 1,
    "balance": 555
}'

> curl --location --request GET 'localhost:8080/api/history/all?sort=date' \
--header 'Authorization: Bearer {some_token}' \
--data-raw ''

> curl --location --request GET 'localhost:8080/api/history/2?limit=2' \
--header 'Authorization: Bearer {some_token}' \
--data-raw ''
