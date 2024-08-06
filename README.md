# Semana tech Rocketseat - GO + REACT

## GENERATE TERN CONFIG AND MIGRATIONS:

tern init ./internal/store/pgstore
tern new --migrations ./internal/store/pgstore create_TABLENAME_table

RUN MIGRATIONS ON GO WRAPPER TO LOAD .ENV FILES

go run cmd/tools/terndotenv/main.go

## Generate SQLC

sqlc generate -f ./internal/store/pgstore/sqlc.yaml

## RUUNNING WITH GO GENERATE

USE ALL GO GENERATE DIRECTIVES ON ALL FOLDERS(./...)
go generate ./...
