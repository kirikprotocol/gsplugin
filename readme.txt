Установка и запуск.
1. Запускаем файл по пути bin/MiniappsTesterGoogleSheetsBot <cfg> передавая вместо <cfg> путь к файлу c json конфигом вида:

{
  "markable": true,
  "serverRoot": "/",
  "port": "32123",
  "unmarkableXML": "C:/Users/gav/go/src/MiniappsTesterGoogleSheetsBot/pages/response.xml",
  "markableXML": "C:/Users/gav/go/src/MiniappsTesterGoogleSheetsBot/pages/markableResponse.xml",
  "logPath": "C:/Users/gav/go/src/MiniappsTesterGoogleSheetsBot/logs/bot.log",
  "pathToGoogleKeyJson": "E:\\passwords\\google_key\\test123-xxxxxxxxxxxx.json",
  "spreadsheetId": "1GCXT5ii2NJxok6hpxxxxxxxxxxxxxxxxxxxxxxxxxx",
  "knownKeys": ["ref_sid", "event.id", "event.order", "subscriber", "abonent", "protocol", "user_id", "service", "event.text", "event.referer", "event", "lang", "serviceId", "wnumber"]
}

По умолчанию он лежит в папке src/cfg.json.

Описание ключей:

markable -- Режим работы - оценивать или нет
serverRoot -- Корневой путь сервера
port -- Порт сервера
unmarkableXML -- Путь к xml ответу сервера в режиме без оценивания(по умолчанию - src/pages/response.xml)
markableXML -- Путь к xml ответу сервера в режиме оценивания(по умолчанию - src/pages/markableResponse.xml)
logPath -- Путь к лог файлу (по умолчанию - src/logs/bot.log)
pathToGoogleKeyJson -- Путь к файлу ключа google для работы с таблицей(подробнее - https://console.developers.google.com/project)
spreadsheetId -- Id таблицы (берется из ее url-a)
knownKeys -- Ключи которые Miniapps передает поумолчанию в запросе

Если ошибок не возникло - значит в конфиге все правильно.

Компиляция.
1. Заходим в папку src
2. Запускаем команду go build .
3. Получаем исходник

Watchdog.
Можно использовать скрипт вида:

#!/bin/bash

DATA_DIR=<путь к распакованному архиву>
PIDFILE=~/.googlesheets_tester_bot_daemon.pid

if [ x"$1" = x-daemon ]; then
  if test -f "$PIDFILE"; then exit; fi
  echo $$ > "$PIDFILE"
  trap "rm '$PIDFILE'" EXIT SIGTERM
  while true; do
        #launch your app here
        $DATADIR/bin/MiniappsTesterGoogleSheetsBot $DATADIR/src/cfg.json
        wait # needed for trap to work
  done
elif [ x"$1" = x-stop ]; then
  kill `cat "$PIDFILE"`
else
  echo "geth daemon: -daemon to start daemon"
  echo "             -stop to stop daemon"
  #echo "             (without any options) -- start daemon"
  exit
  #else
  #nohup "$0" -daemon
fi


Для поддержания бота в рабочем состоянии.

При возникновении внутренних ошибок программы обращаться ко мне - gav@eyeline.mobi
