import type { QuoteRepository } from '@/domain/repositories/QuoteRepository.ts';
import type Quote from '@/domain/types/Quote.ts';

export default class QuoteRepositoryImpl implements QuoteRepository {
  private readonly accessToken: () => string;

  constructor(params: { accessToken: () => string }) {
    this.accessToken = params.accessToken;
  }

  public async getQuote(number: string): Promise<Quote> {
    const response = await fetch(`${import.meta.env.VITE_API}/quotes/${number}`, {
      headers: { Authorization: 'Bearer ' + this.accessToken() }
    });
    return (await response.json()) as Quote;
  }
}
