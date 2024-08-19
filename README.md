![Minecart Rapid Transit Logo](https://github.com/Frumple/mrt-docker-services/assets/68396/32a557d8-f5ad-44ae-9d71-da1ad7d31a55)

# MRT API
A Go-powered API that returns useful data from the [Minecart Rapid Transit (MRT) Minecraft Server](https://www.minecartrapidtransit.net).

## [Swagger (OpenAPI 2.0) Documentation](https://api.minecartrapidtransit.net/swagger/index.html)

## Current Endpoints

- `/warps` - Get warps stored in the [MyWarp](https://github.com/MyWarp/MyWarp) plugin.
- `/companies` - Get companies registered in [this YAML file](https://github.com/Frumple/mrt-api/blob/main/data/companies.yml).
- `/worlds` - Get worlds registered in [this YAML file](https://github.com/Frumple/mrt-api/blob/main/data/worlds.yml).

Note that to ensure performance, the maximum number of warps that can be returned per `/warps` request is **2000**. Use the `offset` query parameter to view warps beyond this limit.

## Example Requests

### Get all warps owned by player "Frumple"
UUID with hyphens:
- `https://api.minecartrapidtransit.net/api/v2/warps?player=ffdaf900-cdb2-4f09-a0fb-81e3087da4e7`

UUID without hyphens:
- `https://api.minecartrapidtransit.net/api/v2/warps?player=ffdaf900cdb24f09a0fb81e3087da4e7`

### Get all warps owned by "West Zeta Rail"
- `https://api.minecartrapidtransit.net/api/v2/warps?company=WZR`

### Get all warps owned by player "FredTheTimeLord" and company "FredRail"
- `https://api.minecartrapidtransit.net/api/v2/warps?player=8ebc51733df2450c92a3e13063409a24&company=FR`

### Get top 10 most visited warps
- `https://api.minecartrapidtransit.net/api/v2/warps?order_by=visits&sort_by=desc&limit=10`

### Get top 10 most visited "IntraRail" warps
- `https://api.minecartrapidtransit.net/api/v2/warps?company=IR&order_by=visits&sort_by=desc&limit=10`

### Get 11th to 20th most visited "IntraRail" warps
- `https://api.minecartrapidtransit.net/api/v2/warps?company=IR&order_by=visits&sort_by=desc&limit=10&offset=10`

### Get 10 newest warps
- `https://api.minecartrapidtransit.net/api/v2/warps?order_by=creation_date&sort_by=desc&limit=10`

### Get 10 oldest "NewRail FLR" warps
- `https://api.minecartrapidtransit.net/api/v2/warps?company=FLR&order_by=creation_date&sort_by=asc&limit=10`

### Get all warps on the Old World
- `https://api.minecartrapidtransit.net/api/v2/warps?world=old`

### Get all private warps
- `https://api.minecartrapidtransit.net/api/v2/warps?type=0`

### Get all companies registered on this API
- `https://api.minecartrapidtransit.net/api/v2/companies`

## Development Setup

Install all dependencies:
```
go build
```

Start the development server:
```
go run .
```

Generate Swagger docs:
```
go install github.com/swaggo/swag/cmd/swag@latest
swag init
```


## License
This API is licensed under the [MIT License](https://choosealicense.com/licenses/mit/).
