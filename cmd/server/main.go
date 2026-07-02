package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/nduagoziem/golang-url-shortener/internal/cache"
	"github.com/nduagoziem/golang-url-shortener/internal/db"
)

func main() {

	ctx := context.Background()

	_ = godotenv.Load()

	db.NewPostgresPool(ctx, os.Getenv("DATABASE_URL"))

	rdb := cache.NewRedisCache(ctx, os.Getenv("REDIS_URL"))
	rdb.Set(ctx, "username", "nduagoziem", 10*time.Minute)
	fmt.Println(rdb.Get(ctx, "username"))
}
