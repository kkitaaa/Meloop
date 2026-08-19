import { Test, TestingModule } from '@nestjs/testing';
import { AppController } from './app.controller';
import { TestRedisService } from './redis/test-redis.service';

describe('AppController', () => {
  let appController: AppController;

  beforeEach(async () => {
    const module: TestingModule = await Test.createTestingModule({
      controllers: [AppController],
      providers: [
        {
          provide: TestRedisService,
          useValue: {
            test: jest.fn().mockResolvedValue('valor'),
          },
        },
      ],
    }).compile();

    appController = module.get<AppController>(AppController);
  });

  describe('redisTest', () => {
    it('should return the Redis test value', async () => {
      const result = await appController.redisTest();

      expect(result).toBe('valor');
    });
  });
});
