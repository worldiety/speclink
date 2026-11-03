<template>
  <p :title="$t('amount')">{{ quote.number }}: {{ quote.amount }}</p>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import type Quote from '@/domain/types/Quote.ts';

defineProps<{ quote: Quote }>();
const { t } = useI18n();
const label = t('number');
</script>
