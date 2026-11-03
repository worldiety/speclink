import type { ThingRepository } from '@/domain/repositories/ThingRepository.ts';
import type { Ok } from '@/domain/types/Ok.ts';

export default class ThingRepositoryImpl implements ThingRepository {
  public async get(): Promise<Ok> {
    return { ok: (await fetch('/ok')).ok };
  }
}
