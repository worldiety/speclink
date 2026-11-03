import type ThingService from '@/domain/services/ThingService.ts';

export default class GoodUseCase {
  constructor(private thingService: ThingService) {}

  execute() {
    this.thingService.audit('canDoThings');
  }
}
