package main

import (
    "context"
    "log"
    "os/signal"
    "syscall"
    "time"

    "github.com/Uva337/WBL0v1/internal/cache"
    "github.com/Uva337/WBL0v1/internal/httpserver"
    "github.com/Uva337/WBL0v1/internal/kafka"
    "github.com/Uva337/WBL0v1/internal/models"
    "github.com/Uva337/WBL0v1/internal/repo"
    "github.com/Uva337/WBL0v1/internal/validator"
)

func main() {
    ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
    defer stop()

    // --- Инициализация валидатора
    v, err := validator.New()
    if err != nil {
        log.Fatalf("failed to init validator: %v", err)
    }

    // --- Инициализация репозитория
    pg, err := repo.NewPostgres(ctx)
    if err != nil {
        log.Fatalf("failed to init postgres: %v", err)
    }
    defer pg.Close()
    log.Println("Postgres connected")

    // --- Инициализация кэша
    c := cache.New(5*time.Minute, 10*time.Minute)
    all, err := pg.GetAll(ctx)
    if err != nil {
        log.Fatalf("failed to get all orders for cache warm up: %v", err)
    }
    if len(all) > 0 {
        c.BulkSet(all)
        log.Printf("cache warmed: %d orders", len(all))
    } else {
        log.Println("cache: no orders in db to warm up")
    }

    // --- Инициализация и запуск Kafka consumer
    cons := kafka.NewConsumer(v)
    defer cons.Close()

    go func() {
        log.Println("Kafka consumer is running...")
        if err := cons.Run(ctx, func(ctx context.Context, o models.Order) error {
            if err := pg.UpsertOrder(ctx, o); err != nil {
                return err
            }
            c.Set(o.OrderUID, o) // Используем Set с UID
            log.Printf("order %s processed and cached", o.OrderUID)
            return nil
        }); err != nil {
            log.Printf("consumer error: %v", err)
        }
    }()

    // --- Инициализация и запуск HTTP-сервера
    srv := httpserver.New(c, pg)
    log.Println("HTTP server is starting...")
    if err := srv.ListenAndServe(ctx); err != nil {
        log.Printf("http server error: %v", err)
    }
}