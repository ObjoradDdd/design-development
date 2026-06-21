# SIMD Reverse Proxy

Легкий reverse proxy на Go с поддержкой:

- сжатия ответов (zstd/gzip) по Accept-Encoding
- кэширования GET/HEAD ответов
- двух типов кэша: in-memory LRU и disk-backed (BoltDB)
- graceful shutdown

## Как это работает

srpc проксирует запросы на target_url, может сжимать ответ и кэшировать его на основе:

- URL и query-параметров
- HTTP-метода
- выбранного Content-Encoding

TTL берется из Cache-Control: max-age=..., а если заголовка нет - из ttl в config.yaml.

## Быстрый старт

1. Поднимите upstream-сервис (по умолчанию ожидается http://localhost:8081).
2. Проверьте config.yaml.
3. Запустите прокси:

go run .

Прокси слушает порт из config.yaml (по умолчанию 8080).

## Пример config.yaml

port: 8080
target_url: "http://localhost:8081"
zstd_level: 3
gzip_level: 5
ttl: 300s
cache_type: "disk"
max_memory_mb: 200
log_level: "info"

## Параметры конфигурации

- port: порт, на котором слушает прокси
- target_url: адрес upstream-сервиса
- zstd_level: уровень сжатия zstd
- gzip_level: уровень сжатия gzip
- ttl: default TTL для кэша (формат duration, например 300s)
- cache_type: тип кэша, disk или любой другой (тогда используется in-memory)
- max_memory_mb: лимит памяти для in-memory кэша
- log_level: уровень логирования (debug, info, warn, error)

## Проверка

Пример запроса через прокси:

curl -i -H "Accept-Encoding: zstd,gzip" "http://localhost:8080/"

## Docker

Сборка и запуск:

docker compose up --build

В docker-compose.yml используется network_mode: host, а config.yaml монтируется в контейнер как /app/config.yaml.
