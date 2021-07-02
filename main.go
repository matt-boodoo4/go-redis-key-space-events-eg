package main

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

func main() {
	// connect to redis
	redisDB := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0,
	})

	_, err := redisDB.Do(context.Background(), "CONFIG", "SET", "notify-keyspace-events", "KEA").Result() // this is telling redis to publish events since it's off by default.
	if err != nil {
		fmt.Printf("unable to set keyspace events %v", err.Error())
		os.Exit(1)
	}
	pubsub := redisDB.PSubscribe(context.Background(), "__keyevent@0__:expired") // this is telling redis to subscribe to events published in the keyevent channel, specifically for expired events

	wg := &sync.WaitGroup{}
	wg.Add(2) // this is for illustration purposes only.

	go func(redis.PubSub) {
		exitLoopCounter := 0
		for { // infinite loop
			message, err := pubsub.ReceiveMessage(context.Background())
			exitLoopCounter++
			if err != nil {
				fmt.Printf("error message - %v", err.Error())
				break
			}
			fmt.Printf("Keyspace event recieved %v  \n", message.String())
			if exitLoopCounter >= 10 {
				wg.Done()
			}
		}
	}(*pubsub)

	go func(redis.Client, *sync.WaitGroup) {
		for i := 0; i <= 10; i++ {
			dynamicKey := fmt.Sprintf("event_%v", i)
			redisDB.Set(context.Background(), dynamicKey, "someval", time.Second*time.Duration(i*2)).Result()
		}
		wg.Done()
	}(*redisDB, wg)

	wg.Wait()
	fmt.Println("exiting program")
}
