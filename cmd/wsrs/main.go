package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"

	"github.com/fekow/semana-tech-go-react-server/internal/api"
	"github.com/fekow/semana-tech-go-react-server/internal/store/pgstore"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		panic(err)
	}

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, fmt.Sprintf(
		"user=%s password=%s host=%s port=%s dbname=%s",
		os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD"), os.Getenv("POSTGRES_HOST"), os.Getenv("POSTGRES_PORT"), os.Getenv("POSTGRES_DB"),
	))
	if err != nil {
		panic(err)
	}

	//defer é antes de retornar a função, roda os defer (tipo um cleanup)
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		panic(err)
	}

	// meu package api tem um metodo NewHandler que retorna um http.Handler
	// o meu sqlc que fez esse pgstore.New
	handler := api.NewHandler(pgstore.New(pool))

	//INICIAR O SERVIDOR HTTP

	go func() {
		if err := http.ListenAndServe(":8080", handler); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				panic(err)
			}
		}
	}()

	// receba o ctrl + c para fechar
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

}
