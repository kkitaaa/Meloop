import { Module } from '@nestjs/common';
import { AppController } from './app.controller';
import { RedisModule } from './redis/redis.module';

@Module({
  imports: [RedisModule],
  controllers: [AppController],
})
export class AppModule {}
