import type { Ok } from '@/domain/types/Ok.ts';

export interface ThingRepository {
  get(): Promise<Ok>;
}
