// K6-TYPES-PURE, twice: an import of a service and a function.
import ThingService from '@/domain/services/ThingService.ts';

export interface Thing {
  service: ThingService;
}

export function helper() {
  return 1;
}
