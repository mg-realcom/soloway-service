# Soloway Service

* gRPC сервер для сбора данных Soloway и записи в BigQuery
* скрипт (клиент) сбора данных по расписанию

## Реализованные методы

* SendReportToStorage - Сбор статистики по площадкам в разрезе дней

Более подробно ознакомиться с протоколами можно в файле `api/grpc/soloway.proto`

## Алгоритм работы

* Получение данных по API Soloway
* Проверка наличия / создание таблицы в BigQuery
* Удаление данных из таблицы BigQuery за период указанный в настройках сбора
* Запись данных в таблицу BigQuery

### Примечания

* BigQuery Dataset должен быть уже создан
* BigQuery Table создается автоматически если отсутствует

## Makefile

`gen_go` - Конверт-ия proto файлов в код Go на основании `api/grpc/soloway.proto`

```bash
make gen
```

`gen_python` - Конверт-ия proto файлов в исходный код Python на основании `api/grpc/soloway.proto`

```bash
make gen_python
```

`build_server` - Компиляция исполняемого файла для сервера (/bin/server_app)

```bash
make build_server
```

`build_schedule_client` - Компиляция исполняемого файла для клиента (/bin/schedule_client)

```bash
make build_schedule_client
```

## Конфигурация (Сервер)

На выбор несколько вариантов настройки:

* По умолчанию в качестве настроек используется `config.yml` в папке с исполняемым файлом
* С помощью флага `-f` укажите путь на конфигурационный файл
* С помощью флага `--env` в качестве настроек используются переменные окружения

### Файл конфигурации

По умолчанию для настроек используется `config.yml` в папке с исполняемым файлом.  
Для использования альтернативного файла используйте флаг `-f`

Шаблон конфигурационного файла находится по пути:  `internal/config/server_config_template.yaml`

```yaml
# Пример конфигурационного файла
keys_dir: "/path/to/keys" // Путь к папке с сервисными ключами
prometheus_addr: localhost:9090

grpc:
  ip: "0.0.0.0"           // Host
  port: 50051             // Порт, который будет прослушивать сервис

tg:
  token: 'TG Token'       // Токен для telegram бота
  chat: 0000000000        // ID чата в который будут отправляться уведомления
  is_enabled: false       // Статус уведомлений
```

### Использование переменных окружения

Для использования переменных окружения используйте флаг  `--env`

| Переменная        | Описание                                         |
|-------------------|--------------------------------------------------|
| `GRPC_IP`         | Host                                             |
| `GRPC_PORT`       | Порт, который будет прослушивать сервис          | 
| `TG_TOKEN`        | Токен для telegram бота                          |
| `TG_CHAT`         | ID чата в который будут отправляться уведомления |
| `TG_ENABLED`      | Статус уведомлений                               |
| `KEYS_DIR `       | Путь к папке с сервисными ключами                |
| `PROMETHEUS_ADDR` | Адрес сервера Prometheus                         |

## Конфигурация (Клиент)

* По умолчанию в качестве настроек используется `schedule_config.yml` в папке с исполняемым файлом
* С помощью флага `-f` укажите путь на конфигурационный файл

### Файл конфигурации

По умолчанию для настроек используется `schedule_config.yml` в папке с исполняемым файлом.  
Для использования альтернативного файла используйте флаг `-f`

Шаблон конфигурационного файла находится по пути:  `internal/config/schedule_config_template.yml`

```yaml
# Пример конфигурационного файла
time: "07:47" // Время ежедневного запуска

grpc:
  ip: "0.0.0.0"           // Host
  port: 50051             // Порт, который будет прослушивать сервис

reports:
  - report_name: "название отчета"
    spreadsheet_id: "google-spreadsheet-id"
    google_service_key: "service_key.json"
    project_id: 'bq-project-id'
    dataset_id: 'bq-dataset-id'
    table_id: 'bq_all_clients_table'
    period: 7

  - report_name: "название отчета 2"
    spreadsheet_id: "google-spreadsheet-id"
    google_service_key: "service_key.json"
    project_id: 'bq-project-id'
    dataset_id: 'bq-dataset-id'
    table_id: 'bq_all_clients_table'
    period: 15
```

## Конфигурация (Google spreadsheet c клиентами)

Лист должен иметь название `// Config`

| Клиент      | Логин         | Пароль           |
|-------------|---------------|------------------|
| client_name | soloway_login | soloway_password |

