# Getapp

Набор инструментов для паблишинга игр и приложений. Вся логика разбита по модулям и сейчас в активной разработке.

![](screens/dashboard.png)

## Установка

Пока нет легкого способа запустить все приложение. Но когда ни будь появиться docker.

Начать нужно с создания базы данных. Все доступы к базе данных нужно указать в конфиге. 
Все конфиги наследуются от `base.yml`
Вы можете завести конфиг с нужным вам именем, например `dev.yml` или `prod.yml` и ужк в этом 
конфиге указывать настройки базы

```yaml
application:
  database:
    host: db.example.ru
    user: db
    password: example
    database: example
```
Для доступа к базе используется библиотека `gorm`. Поэтому, при первом запуске произойдет
автомиграция и создадутся все нужные таблички.

В конфиге можно сразу указать имы и пароль администратора

```yaml
modules:
  admin:
    username: admin
    password: admin
```

Чтобы собрать приложения вам понадобится Go:

```shell
go build -o build/bin/getapp ./cmd/getapp
```

## Запуск

Чтобы запустить приложение нужно выполнить команду:

```shell
./build/bin/getapp -env=dev server
```

На порту 3333 запуститься http сервер. Весь функционал будет доступен по ссылке http://localhost:3333/admin 

## Модули

Сейчас все в разработке. некоторые модули чуть больше готовы к продакшену, а для некоторых пока только заглушки

🔴 - только идея
🟡 - в разработке
🟢 - можно тестить

- 🟡 _boosty_ - модуль, который позволяет использовать boosty как систему подписок в приложении
- 🟢 _billing_ - расширение функционала yoomoney для приема оплат в приложении
- 🔴 _ads_ - рекламный сервер с возможностью кастомной медиации
- 🟢 _lokalize_ - апишка для переводов и локализации приложений
- 🟢 _tracker_ - запись и передача конверсий в рекламные сетки (vk, yandex)
- 🟡 _warehouse_ - простое api для kv хранилища
- 🟡 _landings_ - настройка веб-страниц через базу
- 🔴 _deployer_ - публикация приложений во все сторы и как apk
- 🔴 _codepush_ - обновление приложений без публикации
- 🟡 _mediation_ - медиация инапп рекламы

