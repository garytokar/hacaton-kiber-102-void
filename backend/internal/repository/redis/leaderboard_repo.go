package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type LeaderboardRedisRepo struct {
	client *redis.Client
}

func NewLeaderboardRedisRepo(client *redis.Client) *LeaderboardRedisRepo {
	return &LeaderboardRedisRepo{client: client}
}

const (
	keyAllTimeLeaderboard = "lb:all_time"
	fsmPrefix             = "bot:fsm:"
)

// IncrementUserXP обновляет скор пользователя в глобальном лидерборде Redis
func (r *LeaderboardRedisRepo) IncrementUserXP(ctx context.Context, userID uuid.UUID, xp int) error {
	return r.client.ZIncrBy(ctx, keyAllTimeLeaderboard, float64(xp), userID.String()).Err()
}

// SetBotFSMState сохраняет состояние FSM для Telegram бота
func (r *LeaderboardRedisRepo) SetBotFSMState(ctx context.Context, tgID int64, state string, data string, ttl time.Duration) error {
	pipe := r.client.Pipeline()
	pipe.Set(ctx, fmt.Sprintf("%s%d:state", fsmPrefix, tgID), state, ttl)
	pipe.Set(ctx, fmt.Sprintf("%s%d:data", fsmPrefix, tgID), data, ttl)
	_, err := pipe.Exec(ctx)
	return err
}

func (r *LeaderboardRedisRepo) GetBotFSMState(ctx context.Context, tgID int64) (string, string, error) {
	state, err := r.client.Get(ctx, fmt.Sprintf("%s%d:state", fsmPrefix, tgID)).Result()
	if err == redis.Nil {
		return "", "", nil
	}
	if err != nil {
		return "", "", err
	}
	data, _ := r.client.Get(ctx, fmt.Sprintf("%s%d:data", fsmPrefix, tgID)).Result()
	return state, data, nil
}

func (r *LeaderboardRedisRepo) ClearBotFSM(ctx context.Context, tgID int64) error {
	return r.client.Del(ctx,
		fmt.Sprintf("%s%d:state", fsmPrefix, tgID),
		fmt.Sprintf("%s%d:data", fsmPrefix, tgID),
	).Err()
}
