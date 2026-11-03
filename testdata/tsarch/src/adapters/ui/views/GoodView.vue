<template>
  <p>{{ $t('a') }}</p>
</template>

<script setup lang="ts">
import ServiceLocator from '@/domain/services/ServiceLocator.ts';
import GoodUseCase from '@/domain/usecases/GoodUseCase.ts';

new GoodUseCase(ServiceLocator.getThingService()).execute();
</script>
