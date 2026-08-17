import { Controller, Get } from '@nestjs/common';
import { TestRedisService } from './redis/test-redis.service';

@Controller()
export class AppController {
  constructor(private readonly testRedis: TestRedisService) {}

  @Get('redis-test')
  async redisTest() {
    return await this.testRedis.test();
  }
}
