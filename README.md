# О приложении
### Локальная сборка образа приложения
Для локальной сборки образа приложения достаточно использовать команду ниже, в корне проекта:
```bash
docker build -t alekseeval/logs-scanner:X.Y.Z .
```

### Запуск приложения
Пример команды запуска приложения:
```bash
docker run --network host -d --name logs-scanner alekseeval/logs-scanner:X.Y.Z
```
Ссылка на репозиторий с образами приложения -- https://hub.docker.com/r/alekseeval/logs-scanner/tags

### Конфигурация приложения
Конфигурационный файл приложения должен называться `config.json` и располагаться по пути `/etc/scanner/config.json`

### Swagger
Порт REST API приложения задается в конфигурации -- `system.http.port`. Интерфейс REST API приложения описан в swagger-спецификации, в файле `/swagger/docs-swagger.yaml` 
и доступен по URI `/swagger/`


# Развертывание БД
В этом разделе описано как можно самостоятельно развернуть PostgreSQL БД для проекта

### Развертывание БД запуском скрипта
Для развертывания БД с нуля, достаточно запустить скрипт `db/check_install.sh`. Например, находясь в корне проекта, командой:
```bash
bash db/check_install.sh
```

Предварительно следует передать в скрипт параметры подключения к базе и настройки admin-пользователя.
Сделать это можно изменив непосредственно дефолтные параметры в тексте скрипта:
```txt
15 |  host='192.168.0.108'
16 |  port='5432'
17 |  postgres='postgres'
18 |  postgres_password='postgres'
19 |  db_admin='admin'
20 |  db_admin_password='admin'
21 |  db_name="tool_db"
```
Либо можно задать соответствующие переменные окружения (upper case) и запустить скрипт:
```bash
export HOST='192.168.0.108'
export PORT='5432'
export POSTGRES='postgres'
export POSTGRES_PASSWORD='postgres'
export DB_ADMIN='admin'
export DB_ADMIN_PASSWORD='admin'
export DB_NAME="tool_db"
```

> Переменные окружения, если они установлены, считаются более приоритетными

После исполнения скрипта, логи его работы будут записаны в отдельный файл по пути `db/log/*` и выведены в консоль

### Развертывание БД через Docker-образ

БД можно развернуть запустив Docker-контейнер и передав в него переменные окружения описанные выше.

Ссылка на образ в Docker Hub - TODO: сделать сборку и выложить куда-либо
