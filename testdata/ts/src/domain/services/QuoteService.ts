import type { QuoteRepository } from '@/domain/repositories/QuoteRepository.ts';

export default class QuoteService {
  private readonly quoteRepository: QuoteRepository;

  constructor(params: { quoteRepository: QuoteRepository }) {
    this.quoteRepository = params.quoteRepository;
  }

  public getQuote(number: string) {
    return this.quoteRepository.getQuote(number);
  }
}
