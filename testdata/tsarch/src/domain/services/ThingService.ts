import axios from 'axios';
import GoodUseCase from '@/domain/usecases/GoodUseCase.ts';
import ThingRepositoryImpl from '@/adapters/http/ThingRepositoryImpl.ts';

export default class ThingService {
  public audit(_permission: string) {}

  public run() {
    // K6-DOMAIN-NO-FRAMEWORK: the browser, reached from the domain.
    return [axios, GoodUseCase, ThingRepositoryImpl, localStorage.getItem('x')];
  }
}
