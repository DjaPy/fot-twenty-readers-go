# fot-twenty-readers-go

Go application for generating psalm reading calendars for Orthodox Church readers.

## About the Project

The application creates Excel calendars that distribute the reading of the Psalter (150 psalms divided into 20 kathismas) among 20 readers throughout the year, accounting for the Orthodox calendar.

### Features

- Manage reader groups with customizable start offset
- Generate Excel calendars for any year (2025-2045)
- Store calendars in database
- Retrieve current kathisma by reader number
- Web interface with HTMX

## Quick Start

```bash
# Clone the repository
git clone https://github.com/DjaPy/fot-twenty-readers-go.git
cd fot-twenty-readers-go

# Build and run
go build -o for-twenty-readers cmd/main.go
./for-twenty-readers --port 8080
```

Open browser: http://localhost:8080

## API

### Main Endpoints

```bash
# Create reader group
POST /groups
  name=Church Name&start_offset=1

# Generate calendar
POST /groups/{id}/generate
  year=2025

# Get current kathisma
GET /groups/{id}/current-kathisma?reader_number=5
```

### Example: Get Current Kathisma

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

## Technologies

- **Go 1.22+**
- **Storm/BoltDB** - embedded database
- **Excelize** - Excel generation
- **Chi** - HTTP router
- **HTMX** - dynamic UI
- **Tailwind CSS**

## Development

```bash
just lint           # Linting
just test           # Tests
just all-check      # Full check
```

## Docker

```bash
docker build -t for-twenty-readers .
docker run -p 8080:8080 for-twenty-readers
```

## License

MIT

## Contact

- GitHub: [@DjaPy](https://github.com/DjaPy)
- Issues: [GitHub Issues](https://github.com/DjaPy/fot-twenty-readers-go/issues)