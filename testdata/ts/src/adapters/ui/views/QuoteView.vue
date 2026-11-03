<template>
  <h1>{{ $t('quote') }}</h1>
  <QuoteCard v-if="quote" :quote="quote" />
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue';
import QuoteCard from '@/adapters/ui/components/QuoteCard.vue';
import type Quote from '@/domain/types/Quote.ts';
import ServiceLocator from '@/domain/services/ServiceLocator.ts';
import ShowQuoteUseCase from '@/domain/usecases/ShowQuoteUseCase.ts';

const { number } = defineProps({ number: { type: String, required: true } });
const quote = ref<Quote>();

onMounted(async () => {
  quote.value = await new ShowQuoteUseCase(
    ServiceLocator.getAuthService(),
    ServiceLocator.getQuoteService()
  ).execute(number);
});
</script>
