# ozon-task
### Запуск
```
docker-compose up
```

### Схема проекта
![изображение](https://github.com/JuFnd/ozon-task/assets/109366718/319d945a-f0ab-47ee-8fad-078871b4b602)

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
     
   - Посты:
        Посты реализованы на связке GraphQL + Postgresql/Redis(в зависимости от конфигурации по умолчанию postgres)
        - Схема БД
          
          ![изображение](https://github.com/JuFnd/ozon-task/assets/109366718/2ddaa0ce-ff6f-46d0-99e8-c89d38e4c1b1)

        - СУБД: Postgresql
        - БД Кэширования: Redis

   - Примеры запросов:
     - Регистрация
       ![изображение](https://github.com/JuFnd/ozon-task/assets/109366718/064c2b64-97a3-4e4c-b8a1-47a316d25a20)

     - Авторизация
       ![изображение](https://github.com/JuFnd/ozon-task/assets/109366718/2737acbd-eebc-40e5-ac7f-fe9f17f8a9a6)

     - Выход из аккаунта
       ![изображение](https://github.com/JuFnd/ozon-task/assets/109366718/29320114-ac6d-4b80-a09b-f8e161a3d45a)

     - Создание постов с конфигурированием комментариев(разрешены/запрещены)
     ![bkg114-k9gE](https://github.com/JuFnd/ozon-task/assets/109366718/165b616c-e4a4-4c96-a3cd-89cf8b8e2390)
     ![LGRcAuTlJho](https://github.com/JuFnd/ozon-task/assets/109366718/c3993618-513d-4ca3-9ae0-00460595eaef)

     - Написание комментария под постом(разрешены/запрещены)
     ![w76yKNlQ7No](https://github.com/JuFnd/ozon-task/assets/109366718/c2b780f6-f621-4691-9426-34b51e45dee0)
     ![EvbKB3z1vTQ](https://github.com/JuFnd/ozon-task/assets/109366718/ff30d116-4f6e-4ff0-8050-817d9a1f5cda)

     - Получение поста
     ![PzGwm9UFCeo](https://github.com/JuFnd/ozon-task/assets/109366718/d6b7b22c-5887-4036-b800-8a14c68ab022)

     - Получение комментариев
     ![JFnUx9JNCT8](https://github.com/JuFnd/ozon-task/assets/109366718/c75723c0-a2f0-4c48-a5fa-9133a594e488)






