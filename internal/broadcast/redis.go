package broadcast

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/redis/go-redis/v9"
)

const Channel = "blockscan:events"

type Publisher struct{
	client *redis.Client
}

func NewPublisher()*Publisher{
	Redis_Url := os.Getenv("REDIS_URL")
	return &Publisher{
		client: redis.NewClient(&redis.Options{
			Addr: Redis_Url,
		}),
	}
}

func(p *Publisher) Publish(ctx context.Context,eventType string, data any)error{
	payload := map[string]any{"type":eventType,"data":data}

	b,err:= json.Marshal(payload)

	if err !=nil{
		return fmt.Errorf("marshal: %w",err)

	}
	return p.client.Publish(ctx,Channel,b).Err()
}