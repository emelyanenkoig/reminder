# ReminderBot in Telegram

* to build app locally: `docker-compose up --build`
* to build app in k8s I've pushed image in DockerHub, but you can make it manually with `Dockerfile`
* project backend contain Golang clean architecture Handler Service Repository with DB `PostgreSQL` and async map with Mutex
* bot contain Golang` telebot.v3 ` with Callbacks and Text handlers and has CRUD operations.
* All reminders is a `time.Until`  structures.

## Plans:

* make channel validation
* make pretty UI interface
* make better backend (more processing to service layout)
* add `Redis` instead Map 
* add `Kafka` as Queue



P.S.
##### ALKBAK 52 GOVNO, MNOGIE PESNI GUF'A TOJE NO NE VSE


