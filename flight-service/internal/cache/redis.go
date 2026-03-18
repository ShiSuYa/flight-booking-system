package cache

import (
    "context"
    "log"
    "time"

    "github.com/redis/go-redis/v9"
)

var Ctx = context.Background()

func NewRedisClient() *redis.Client {
    var rdb *redis.Client
    var err error

    for i := 0; i < 10; i++ {
        rdb = redis.NewFailoverClient(&redis.FailoverOptions{
            MasterName:    "mymaster",
            SentinelAddrs: []string{"redis-sentinel:26379"},
            DB:            0,
        })

        _, err = rdb.Ping(Ctx).Result()
        if err == nil {
            log.Println("Connected to Redis via Sentinel")
            return rdb
        }

        log.Printf("Redis not ready, retrying... (%d/10)", i+1)
        time.Sleep(2 * time.Second)
    }

    log.Fatalf("Failed to connect to Redis after retries: %v", err)
    return nil
}	