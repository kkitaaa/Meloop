import { Module } from '@nestjs/common';
import { createClient } from 'redis';
import { TestRedisService } from './test-redis.service';

const redisProvider = {
  provide: 'REDIS_CLIENT',
  useFactory: async () => {
    const client = createClient({
      url: 'redis://:{*{Kt_|5vIUk>>{;3wAe@localhost:6379',
    });
    await client.connect();
    return client;
  },
};

@Module({
  providers: [redisProvider, TestRedisService],
  exports: ['REDIS_CLIENT', TestRedisService],
})
export class RedisModule {}
