import ServiceLocator from '@/domain/services/ServiceLocator.ts';
import GoodUseCase from '@/domain/usecases/GoodUseCase.ts';

export function setup() {
  new GoodUseCase(ServiceLocator.init());
}
