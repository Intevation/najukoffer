# README

- [README](#readme)
  - [Requirements](#requirements)
  - [Building on file save](#building-on-file-save)
  - [JWT](#jwt)
    - [Generate token with jwt-generator](#generate-token-with-jwt-generator)
    - [Get data](#get-data)
  - [Build on save](#build-on-save)
  - [mysql snippets](#mysql-snippets)

## Requirements

```shell
go get -u -v github.com/go-sql-driver/mysql
go get -u -v github.com/dgrijalva/jwt-go
go get -u -v github.com/go-sql-driver/mysql
go get -u -v github.com/joho/godotenv
go get -u -v github.com/paulmach/go.geojson
go get -u -v github.com/rs/cors
```

## Building on file save

```shell
while inotifywait -e close_write *.go; do go build ; done
```

## JWT

### Generate token with jwt-generator

- https://github.com/rybit/jwt-generator

```shell
jwt-generator gen -s secret -u user
```

- `-s` the secret / signingKey
- `-u` the user to have for the token

### Get data

```shell
curl -H "Token: XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX" localhost:3000/this_week
```

## Build on save

```shell
while inotifywait -e close_write *.go; do go build ; done
```

## mysql snippets

```sql
SELECT DATE_ADD(curdate(), INTERVAL 6 MONTH);
```

```sql
SELECT thema,DATE_FORMAT(von,'%d.%m.%Y %H:%i') as von,DATE_FORMAT(bis,'%d.%m.%Y %H:%i') as bis,typ,plz,ort,bundesland,beschreibung,eingetragen_von,eingetragen_von_kontakt,X,Y FROM dates_with_location WHERE date(von) between curdate() and DATE_ADD(curdate(), INTERVAL 6 MONTH);
```

```sql
SELECT thema,DATE_FORMAT(von,'%d.%m.%Y %H:%i') as von,DATE_FORMAT(bis,'%d.%m.%Y %H:%i') as bis,typ,plz,ort,bundesland,beschreibung,eingetragen_von,eingetragen_von_kontakt,X,Y FROM dates_with_location WHERE date(von) between curdate() and DATE_ADD(curdate(), INTERVAL 6 MONTH) and TYP REGEXP 'NAJU';
```

```sql
CREATE OR REPLACE VIEW next_6month AS SELECT thema,DATE_FORMAT(von,'%d.%m.%Y %H:%i') as von,DATE_FORMAT(bis,'%d.%m.%Y %H:%i') as bis,typ,plz,ort,bundesland,beschreibung,eingetragen_von,eingetragen_von_kontakt,X,Y FROM dates_with_location WHERE date(von) between curdate() and DATE_ADD(curdate(), INTERVAL 6 MONTH);
```
