package gofactory

import (
	"context"
	"errors"
	"time"
)

func (s *Service) RedisReadKeys(key string) ([]string, error) {
	err := s.RedisClientLoaded()
	if err != nil {
		return []string{}, err
	}
	ss, err := s.opt.cliredis.keys(key)
	if s.checkRedisDialErr(err) != nil {
		return []string{}, err
	}
	s.opt.logg.Debug("[redis] read keys:" + key)
	return ss, nil
}

func (s *Service) RedisReadHashField(key, field string) (string, error) {
	err := s.RedisClientLoaded()
	if err != nil {
		return "", err
	}
	ctx, cancel := context.WithTimeout(context.Background(), s.opt.cliredis.readTimeout)
	defer cancel()
	val := s.opt.cliredis.cli.HGet(ctx, key, field)
	if s.checkRedisDialErr(val.Err()) != nil {
		return "", val.Err()
	}
	s.opt.logg.Debug("[redis] read hash field:" + key)
	return val.Val(), nil
}

func (s *Service) RedisReadHashMap(key string) (map[string]string, error) {
	err := s.RedisClientLoaded()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), s.opt.cliredis.readTimeout)
	defer cancel()
	val := s.opt.cliredis.cli.HGetAll(ctx, key)
	if s.checkRedisDialErr(val.Err()) != nil {
		return nil, val.Err()
	}
	s.opt.logg.Debug("[redis] read hash:" + key)
	return val.Val(), nil
}

func (s *Service) RedisRead(key string) (string, error) {
	err := s.RedisClientLoaded()
	if err != nil {
		return "", err
	}
	val, err := s.opt.cliredis.read(key)
	// ctx, cancel := context.WithTimeout(context.Background(), s.opt.cliredis.readTimeout)
	// defer cancel()
	// ans := s.opt.cliredis.cli.Get(ctx, key)
	// err = s.checkRedisDialErr(ans.Err())
	if err != nil {
		return "", err
	}
	s.opt.logg.Debug("[redis] read:" + key)
	return val, nil
}

func (s *Service) RedisDelKey(key string) error {
	err := s.RedisClientLoaded()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), s.opt.cliredis.writeTimeout)
	defer cancel()
	err = s.checkRedisDialErr(s.opt.cliredis.cli.Del(ctx, key).Err())
	if err != nil {
		return err
	}
	s.opt.logg.Debug("[redis] delete key:" + key)
	return nil
}

func (s *Service) RedisDelHashField(key, field string) error {
	err := s.RedisClientLoaded()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), s.opt.cliredis.writeTimeout)
	defer cancel()
	err = s.checkRedisDialErr(s.opt.cliredis.cli.HDel(ctx, key, field).Err())
	if err != nil {
		return err
	}
	s.opt.logg.Debug("[redis] delete key:" + key)
	return nil
}

func (s *Service) RedisExpireKey(key string, expire time.Duration) error {
	err := s.RedisClientLoaded()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), s.opt.cliredis.writeTimeout)
	defer cancel()
	err = s.checkRedisDialErr(s.opt.cliredis.cli.Expire(ctx, key, expire).Err())
	if err != nil {
		return err
	}
	s.opt.logg.Debug("[redis] expire key:" + key)
	return nil
}

func (s *Service) RedisWriteHashMap(key string, value map[string]any) error {
	if len(value) == 0 {
		return nil
	}
	err := s.RedisClientLoaded()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), s.opt.cliredis.writeTimeout)
	defer cancel()
	if s.opt.cliredis.cliver < 4 {
		args := make([]any, 0, len(value)*2)
		for f, v := range value {
			args = append(args, f, v)
		}
		err = s.checkRedisDialErr(s.opt.cliredis.cli.HMSet(ctx, key, args...).Err())
	} else {
		err = s.checkRedisDialErr(s.opt.cliredis.cli.HSet(ctx, key, value).Err())
	}
	if err != nil {
		return err
	}
	s.opt.logg.Debug("[redis] write hash:" + key)
	return nil
}

func (s *Service) RedisWriteHashField(key, field string, value any) error {
	return s.RedisWriteHashMap(key, map[string]any{field: value})
}

func (s *Service) RedisWrite(key string, value any, expire time.Duration) error {
	err := s.RedisClientLoaded()
	if err != nil {
		return err
	}
	err = s.opt.cliredis.write(key, value, expire)
	// ctx, cancel := context.WithTimeout(context.Background(), s.opt.cliredis.writeTimeout)
	// defer cancel()
	// err = s.checkRedisDialErr(s.opt.cliredis.cli.Set(ctx, key, value, expire).Err())
	if s.checkRedisDialErr(err) != nil {
		return err
	}
	s.opt.logg.Debug("[redis] write:" + key)
	return nil
}

func (s *Service) RedisClientLoaded() error {
	ok := s.opt.cliredis.loaded.Load()
	if ok {
		return nil
	}
	s.opt.logg.Error("[redis] client not loaded")
	return errors.New("[redis] client not loaded")
}

func (s *Service) checkRedisDialErr(err error) error {
	if err == nil {
		return nil
	}
	s.opt.logg.Error("[redis] error:" + s.opt.cliredis.checkRedisDialErr(err).Error())
	return err
}
