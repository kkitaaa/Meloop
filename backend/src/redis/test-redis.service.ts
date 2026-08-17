import { Inject, Injectable } from '@nestjs/common';
import { createClient } from 'redis';

type RedisClient = ReturnType<typeof createClient>;

@Injectable()
export class TestRedisService {
  constructor(@Inject('REDIS_CLIENT') private readonly redis: RedisClient) {}

  async test() {
    await this.redis.set('clave', 'valor');
    const result = await this.redis.get('clave');
    return result;
  }
}
