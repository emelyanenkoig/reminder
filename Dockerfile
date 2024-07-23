# Используем официальный образ Go для сборки
FROM golang:1.22.1-alpine AS builder

# Устанавливаем необходимые зависимости
RUN apk update && apk add --no-cache git

# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /app

# Копируем файлы go.mod и go.sum и загружаем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем остальной исходный код приложения
COPY . .

# Сборка Go приложения
RUN GOARCH=amd64 GOOS=linux go build -o reminder .

# Используем минимальный образ для запуска приложения
FROM alpine:latest

# Устанавливаем сертификаты и другие зависимости времени выполнения
RUN apk --no-cache add ca-certificates

# Устанавливаем рабочую директорию для runtime образа
WORKDIR /root/

# Копируем скомпилированный бинарный файл из стадии сборки
COPY --from=builder /app/reminder .

# Открываем порт, на котором приложение будет слушать
EXPOSE 8080

# Команда для запуска приложения
CMD ["./reminder"]
