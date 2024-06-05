# ozon-task
### Описание архитектуры:
   Реализована микросевисная архитектура, общение сервисов по gRPC.
   - Реализовано два микросервиса:
   - Авторизация:
        Авторизация реализована на основе сессий.
        ![UNXyORX2pQ8](https://github.com/JuFnd/avito-task/assets/109366718/0a8f1eaa-9af5-4eef-bfc2-df2969b1bc46)

        - Схема БД:

          ![изображение](https://github.com/JuFnd/avito-task/assets/109366718/a36e0419-5f02-4d8d-a069-87d5304ffafd)

        - СУБД: Postgresql
        - БД Кэширования: Redis
