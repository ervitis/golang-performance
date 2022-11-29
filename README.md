# Performance in go

Read multiple CSV files with the following header data:

- id (int)
- first_name (string)
- last_name (string)
- email (string)
- gender (string)
- country (string)
- birthday (date YYYY/MM/DD)

## HowTo

### Requirements:

- Docker
- Golang 1.17 up
- Docker compose

### Execution

```bash
docker-compose up --build
```

Access to Grafana page:

> localhost:3000

Import the dashboard from the file in `infra/metrics/grafana/dashboards/panels.json` into Grafana
