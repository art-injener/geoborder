## Сервис Geoborder X-Keeper
  Поиск вхождения точек в геозоны, расчет дистанции до границы геозоны

    ### Сборка docker образа

1. Настройка файла конфигурации запуска

   В корневой директории проект в папке configs лежит файл `app.env.example`. в нем приведен список всех настроек для запуска приложения

```yaml
    LOG_LEVEL = debug - уровень вывода логов

    DEVICES_DB_HOST=devices
    DEVICES_DB_PORT=5432
    DEVICES_DB_USERNAME=x-keeper
    DEVICES_DB_PASSWORD=1234567
    DEVICES_DB_DATABASE=x-keeper_devices
    
    
    GRPS_PORT = 6589 - порт для запуска сервера gRPC 
```

Создаем копию этого файла в папке configs. Переименовываем его в app.env, заполняем параметрами подключения

2. Выполняем сборку образа:
    - выполнив в консоле команду `docker build  -t api_service .`
3. Запуск приложения :
    - выполнив в консоле команду `docker run  --network=host --restart=always -d api_service`

