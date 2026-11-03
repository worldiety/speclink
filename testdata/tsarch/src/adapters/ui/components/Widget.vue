<template>
  <p>{{ $t('missingKey') }}</p>
</template>

<script setup lang="ts">
import ThingRepositoryImpl from '@/adapters/http/ThingRepositoryImpl.ts';
import type { ThingRepository } from '@/domain/repositories/ThingRepository.ts';
import ThingService from '@/domain/services/ThingService.ts';
import ServiceLocator from '@/domain/services/ServiceLocator.ts';

const things: ThingRepository = new ThingRepositoryImpl();
const service: ThingService = ServiceLocator.getThingService();
console.log(things, service);
</script>
