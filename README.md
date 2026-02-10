# Простой рекламный сервер на GoLang

Посмотреть видео по проекту можете тут https://youtu.be/27WPASOQs2w

В этом репозитории находится простейший рекламный сервер на Go. Несмотря на его простоту, он использует стек, который используется на самом настоящем высоконагруженном рекламном сервере: fasthttp, GeoIP от MaxMind и [mssola/user_agent](https://github.com/mssola/user_agent).

## Обновления!

* Совсем недавно вышло видео про __[реализацию высоконагруженной статистики](https://www.youtube.com/watch?v=eAmblLikTgo)__

  * [Ветка со статистикой на MySQL](https://github.com/VyacheArt/simple-ads-server/tree/mysql-stats)
  * [Ветка со статистикой на ClickHouse](https://github.com/VyacheslavGoryunov/simple-ads-server/tree/clickhouse-stats)

* А в ролике про __[Prometheus метрики в GoLang](https://youtu.be/6pQQw-qEoCo)__ я рассказываю про метрики на примере этого рекламного сервера
  * [Ветка с метриками на Prometheus](https://github.com/VyacheArt/simple-ads-server/tree/prometheus)

## Настройка

Не забудьте перед запуском положить GeoLite2-Country.mmdb в корень проекта. В остальном всё должно сразу работать :) 

## Стек

* Быстрый HTTP сервер с использованием [valyala/fasthttp](https://github.com/valyala/fasthttp)
* Парсинг User-Agent через [mssola/user_agent](https://github.com/mssola/user_agent)
* Взаимодействие с GeoIP через [oschwald/geoip2-golang](https://github.com/oschwald/geoip2-golang)

grafana
pass: admin1


#### Как он работает

```
1. Данные попадают в память при каждом запросе
В HTTP-обработчике при каждом запросе вызывается:
server.go
Lines 69-69
    s.stats.Append(key, value)
Append кладёт/агрегирует пару ключ–значение в in-memory map m.rows внутри stats.Manager (под мьютексом).
2. Loop работает только по таймеру
В stats.Manager loop выглядит так:
manager.go
Lines 79-90
// loop периодически инициирует запись статистики или завершает работу по ctx.func (m *Manager) loop() {    for {        select {        case <-time.After(m.flushInterval):            m.startInserting()        case <-m.ctx.Done():            m.startInserting()            return        }    }}
Никаких каналов с «новыми данными» нет: раз в flushInterval (у тебя 10 секунд) срабатывает time.After, и вызывается m.startInserting().
3. Что делает startInserting
Вызывает withdraw() — атомарно забирает текущий map m.rows и подменяет его новым пустым.
Если после withdraw() map пустой — логирует "No stats rows, skipping" и выходит.
Если не пустой — отдаёт эти строки в m.writer.Insert(rows), то есть в ClickHouse.
То есть каждые 10 секунд в ClickHouse уходит всё, что успело накопиться в m.rows с прошлого flush’а. Новые строки — это просто то, что успели добавить вызовы s.stats.Append(key, value) между двумя срабатываниями таймера.
Схема потока
Запрос 1 → stats.Append(...)  → m.rows[key1] = value1Запрос 2 → stats.Append(...)  → m.rows[key2] = value2  (или агрегация по ключу)...через 10 сек → time.After → startInserting() → withdraw() забирает весь m.rows            → writer.Insert(rows) → INSERT в ClickHouse            → m.rows заменён на пустой map, накопление идёт заново
Итого: появление новых rows loop не отслеживает — он просто периодически «выгребает» всё из m.rows и пишет в ClickHouse. Новые строки — это то, что добавилось в этот map между двумя такими срабатываниями.1. Данные попадают в память при каждом запросе
В HTTP-обработчике при каждом запросе вызывается:
server.go
Lines 69-69
    s.stats.Append(key, value)
Append кладёт/агрегирует пару ключ–значение в in-memory map m.rows внутри stats.Manager (под мьютексом).
2. Loop работает только по таймеру
В stats.Manager loop выглядит так:
manager.go
Lines 79-90
// loop периодически инициирует запись статистики или завершает работу по ctx.func (m *Manager) loop() {    for {        select {        case <-time.After(m.flushInterval):            m.startInserting()        case <-m.ctx.Done():            m.startInserting()            return        }    }}
Никаких каналов с «новыми данными» нет: раз в flushInterval (у тебя 10 секунд) срабатывает time.After, и вызывается m.startInserting().
3. Что делает startInserting
Вызывает withdraw() — атомарно забирает текущий map m.rows и подменяет его новым пустым.
Если после withdraw() map пустой — логирует "No stats rows, skipping" и выходит.
Если не пустой — отдаёт эти строки в m.writer.Insert(rows), то есть в ClickHouse.
То есть каждые 10 секунд в ClickHouse уходит всё, что успело накопиться в m.rows с прошлого flush’а. Новые строки — это просто то, что успели добавить вызовы s.stats.Append(key, value) между двумя срабатываниями таймера.
Схема потока
Запрос 1 → stats.Append(...)  → m.rows[key1] = value1Запрос 2 → stats.Append(...)  → m.rows[key2] = value2  (или агрегация по ключу)...через 10 сек → time.After → startInserting() → withdraw() забирает весь m.rows            → writer.Insert(rows) → INSERT в ClickHouse            → m.rows заменён на пустой map, накопление идёт заново
Итого: появление новых rows loop не отслеживает — он просто периодически «выгребает» всё из m.rows и пишет в ClickHouse. Новые строки — это то, что добавилось в этот map между двумя такими срабатываниями.
```