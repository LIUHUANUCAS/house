# House Data API

GET API PATH: `/v1/daily_house`
POST API PATH: `add_daily_house`
protocol: json

Data model:

```go
{
    total_count int
    total_area float
    house_count int
    house_area float
}

new house: day="2025-06-09-08"

```
