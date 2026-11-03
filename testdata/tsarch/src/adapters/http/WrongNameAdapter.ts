import type { ThingRepository } from '@/domain/repositories/ThingRepository.ts';
import type { Ok } from '@/domain/types/Ok.ts';

export class WrongNameAdapter implements ThingRepository {
  public async get(): Promise<Ok> {
    return { ok: true };
  }
}
