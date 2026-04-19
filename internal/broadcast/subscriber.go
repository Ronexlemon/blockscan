package broadcast

import (
	"context"
	"log"

	"github.com/ronexlemon/blockscan/internal/api"
)




func (r *Publisher) StartSubscriber(ctx context.Context,hub *api.SSEHub){
	sub := r.client.Subscribe(ctx,Channel)

	go func ()  {
		defer sub.Close()
		log.Println("Redis subscribe")

		ch := sub.Channel()
		for{
			select{
			case msg := <-ch:
				hub.BroadcastRaw([]byte(msg.Payload))
			case <-ctx.Done():
				log.Println("Redis subscriber stopped")
				return
			}
		}
		
	}()
}



