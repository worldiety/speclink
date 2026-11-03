import type AuthService from '@/domain/services/AuthService.ts';
import type QuoteService from '@/domain/services/QuoteService.ts';

export default class ShowQuoteUseCase {
  constructor(
    private authService: AuthService,
    private quoteService: QuoteService
  ) {}

  /**
   * Loads a quote to show it.
   */
  execute(number: string) {
    this.authService.audit('canViewQuotes');
    return this.quoteService.getQuote(number);
  }
}
