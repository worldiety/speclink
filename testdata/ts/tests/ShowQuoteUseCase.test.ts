import { describe, expect, test } from 'vitest';
import { verified } from '@/speclink.ts';
import { RQuoteShow } from '../requirements/fun/quote/R-QUOTE-SHOW.req.ts';
import { RDecHexagonal } from '../requirements/dec/R-DEC-HEXAGONAL.req.ts';
import ShowQuoteUseCase from '@/domain/usecases/ShowQuoteUseCase.ts';
import AuthService from '@/domain/services/AuthService.ts';
import QuoteService from '@/domain/services/QuoteService.ts';

describe('ShowQuoteUseCase', () => {
  test('shows number and amount', async () => {
    const quote = await new ShowQuoteUseCase(
      new AuthService({ permissions: { canViewQuotes: true } }),
      new QuoteService({ quoteRepository: { getQuote: async (n) => ({ number: n, amount: 1 }) } })
    ).execute('A-1');
    expect(quote).toEqual({ number: 'A-1', amount: 1 });
    verified(RQuoteShow, RDecHexagonal);
  });

  test('refuses without permission', async () => {
    await expect(async () =>
      new ShowQuoteUseCase(
        new AuthService({ permissions: { canViewQuotes: false } }),
        new QuoteService({ quoteRepository: { getQuote: async (n) => ({ number: n, amount: 1 }) } })
      ).execute('A-1')
    ).rejects.toThrow();
  });
});
