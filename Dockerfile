# Используем официальный образ Go для сборки приложения
FROM golang:1.22.1-alpine AS builder

# Установим зависимости
RUN apk update && apk add --no-cache git

# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /app

# Копируем файлы в контейнер
COPY . .

# Сборка Go-приложения
RUN go mod tidy
RUN go build -o reminder .

# Используем минимальный образ для запуска собранного приложения
FROM alpine:latest

# Установим сертификаты
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Копируем собранное приложение из предыдущего этапа
COPY --from=builder /app/reminder .

# Копируем файл конфигурации
COPY pkg/config/config.yaml /root/pkg/config/config.yaml

# Экспонируем порт, который используется приложением
EXPOSE 8080

# Команда для запуска приложения
CMD ["./reminder"]
