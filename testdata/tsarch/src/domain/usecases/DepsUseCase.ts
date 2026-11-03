import type { ThingRepository } from '@/domain/repositories/ThingRepository.ts';

export default class DepsUseCase {
  constructor(private things: ThingRepository) {}

  execute(service: { audit(p: string): void }) {
    service.audit('x');
    return this.things;
  }
}
