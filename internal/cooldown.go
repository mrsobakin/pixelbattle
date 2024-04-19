package internal

import (
	"github.com/mrsobakin/pixelbattle/auth"
	"sync"
	"time"
)

type CooldownManager struct {
	lock      sync.Mutex
	cooldowns map[auth.UserId]time.Time
	Cooldown  time.Duration
}

func NewCooldownManager(cooldown time.Duration) *CooldownManager {
	return &CooldownManager{
		cooldowns: make(map[auth.UserId]time.Time),
		Cooldown:  cooldown,
	}
}

func (cm *CooldownManager) Attempt(id auth.UserId) (bool, *time.Duration) {
	cm.lock.Lock()
	defer cm.lock.Unlock()

	last, has := cm.cooldowns[id]

	if !has || time.Since(last) > cm.Cooldown {
		cm.cooldowns[id] = time.Now()
		return true, nil
	}

	remains := cm.Cooldown - time.Since(last)

	return false, &remains
}
