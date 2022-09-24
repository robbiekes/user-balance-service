package rediscache

import "time"

type Option func(r *Redis)

func Expire(expire time.Duration) Option {
	return func(r *Redis) {
		r.expire = expire
	}
}
