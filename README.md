# Eginx - load-balancer and rate-limiter
*[cloud.ru camp spring 2025](https://github.com/Go-Cloud-Camp/test-assignment)*

## Компоненты 

- cmd/api/main.go - API для создания/удаления клиентов
- сmd/app/main.go - Load-balancer и rate-limiter
- cmd/backend/main.go - Пример backend-сервера

## Конфигурация 

Для запуска необходимо указать конфигурационный файл в формате JSON. Пример:

```json
{
    "version": "0.0.1",                 // Версия конфигурации
    "port": 8080,                       // Порт для запуска load-balancer
    "targets": [                        // Список для проксирования, можно изменять не выключая load-balancer   
        "http://localhost:5000",
        "http://localhost:5001",
        "http://localhost:5002",
        "http://localhost:5003"
    ],
    "limiter": {                        // Настройки rate-limiter
        "enabled": true,                // Включение rate-limiter
        "defaultRPM": 4                 // Количество запросов в минуту для всех клиентов
    },
    "redis": {                          // Настройки redis
        "host": "localhost",            // Хост redis
        "port": 6379,                   // Порт redis
        "password": ""                  // Пароль redis
    }
}
```


## Сборка 


### Ручная сборка
Для сборки необходимо выполнить команду:


```bash
go build -o eginx ./cmd/app/main.go         # Load-balancer
go build -o backend ./cmd/backend/main.go   # Backend
go build -o api ./cmd/api/main.go           # API
```

### Makefile

```bash
make build-all # Сборка всех компонентов в директорию bin
```


## Запуск 

Для запуска необходимо выполнить команду:

```bash
./bin/eginx -config ./path/to/config.json           # Load-balancer
./bin/api -config ./path/to/config.json             # API
./bin/backend -port {port}                          # Backend
```
**Примечание:**
- если не указан параметр -config, то будет использован файл config.json из текущей директории 
- если не указан параметр -port, то будет использован порт 5000 (если порт занят, то backend будет инкрементироваться до доступного)

### Docker

Для звпуска в docker необходимо выполнить команду:

```bash
docker compose up -d --build
```

Команда запустит 
- Load-balancer (1 instance) - 8080 порт
- API (1 instance) - 7000 порт
- Backend (5 instances) - 5000-5004 порты


## Тесты

[Результат Apache Bench (n=5000 c=1000)](./tests/apache-bench/n5000c1000.log)