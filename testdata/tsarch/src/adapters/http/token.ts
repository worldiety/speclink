import ServiceLocator from '@/domain/services/ServiceLocator.ts';

export const token = () => ServiceLocator.getThingService();
