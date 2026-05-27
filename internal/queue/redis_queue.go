package queue
import(
	"context"
	"github.com/redis/go-redis/v9"
)

type RedisQueue struct{
	client *redis.Client
}

func NewRedisQueue(client *redis.Client) *RedisQueue{
	return &RedisQueue{
		client: client,
	}
}

func (q *RedisQueue) Enqueue (ctx context.Context, queueName string, payload []byte) error {
	return q.client.LPush(ctx, queueName, payload).Err()
}

func (q *RedisQueue) Dequeue (ctx context.Context, queuName string) ([]byte, error) {
	result, err := q.client.BRPop(ctx,0,queuName).Result()
	if err != nil {
		return nil, err
	}
	return []byte(result[1]), nil
}
