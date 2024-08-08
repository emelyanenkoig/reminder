# Этап сборки
FROM golang:1.22.1-alpine AS builder

# Устанавливаем рабочую директорию для сборки
WORKDIR /app

# Копируем go mod и go sum файлы
COPY go.mod go.sum ./

# Загружаем зависимости
RUN go mod download

# Копируем остальной исходный код приложения
COPY . .

# Сборка Go приложения
RUN GOARCH=amd64 GOOS=linux go build -o reminder .

# Используем минимальный образ для запуска приложения
FROM alpine:latest

# Устанавливаем сертификаты и временную зону
RUN apk --no-cache add ca-certificates tzdata

# Настраиваем временную зону на Moscow
ENV TZ=Europe/Moscow

# Устанавливаем рабочую директорию для runtime образа
WORKDIR /root/

# Копируем скомпилированный бинарный файл из стадии сборки
COPY --from=builder /app/reminder .

# Открываем порт, на котором приложение будет слушать
EXPOSE 8080

# Команда для запуска приложения
CMD ["./reminder"]

