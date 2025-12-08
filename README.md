# fot-twenty-readers-go

Go-приложение для генерации календарей чтения псалмов для чтецов.

## О проекте

Приложение создает Excel-календари, распределяющие чтение Псалтири (150 псалмов, разделенных на 20 кафизм) между 20 чтецами в течение года с учетом православного календаря.

### Возможности

- Управление группами чтецов с настраиваемым стартовым смещением
- Генерация Excel-календарей на любой год (2025-2045)
- Хранение календарей в базе данных
- Получение текущей кафизмы по номеру чтеца
- Web-интерфейс с HTMX

## Быстрый старт

```bash
# Клонировать репозиторий
git clone https://github.com/DjaPy/fot-twenty-readers-go.git
cd fot-twenty-readers-go

# Собрать и запустить
go build -o for-twenty-readers cmd/main.go
./for-twenty-readers --port 8080
```

Откройте браузер: http://localhost:8080

## API

### Основные endpoints

```bash
# Создать группу чтецов
POST /groups
  name=Храм Покрова&start_offset=1

# Сгенерировать календарь
POST /groups/{id}/generate
  year=2025

# Узнать текущую кафизму
GET /groups/{id}/current-kathisma?reader_number=5
```

### Пример: узнать текущую кафизму

```bash
curl "http://localhost:8080/groups/{group-id}/current-kathisma?reader_number=5"
```

Response:
```json
{
  "reader_number": 5,
  "date": "2025-12-07",
  "kathisma": 19,
  "year": 2025
}
```

## Технологии

- **Go 1.22+**
- **Storm/BoltDB** - встроенная БД
- **Excelize** - генерация Excel
- **Chi** - HTTP роутер
- **HTMX** - динамический UI
- **Tailwind CSS**

## Разработка

```bash
just lint           # Линтинг
just test           # Тесты
just all-check      # Полная проверка
```

## Docker

```bash
docker build -t for-twenty-readers .
docker run -p 8080:8080 for-twenty-readers
```

## Лицензия

MIT

## Контакты

- GitHub: [@DjaPy](https://github.com/DjaPy)
- Issues: [GitHub Issues](https://github.com/DjaPy/fot-twenty-readers-go/issues)