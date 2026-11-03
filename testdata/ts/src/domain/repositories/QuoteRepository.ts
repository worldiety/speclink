import type Quote from '@/domain/types/Quote.ts';

export interface QuoteRepository {
  getQuote(number: string): Promise<Quote>;
}
