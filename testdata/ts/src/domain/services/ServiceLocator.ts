import type AuthService from './AuthService';
import type QuoteService from './QuoteService';

export default class ServiceLocator {
  private static authService: AuthService;
  private static quoteService: QuoteService;

  public static init(authService: AuthService, quoteService: QuoteService) {
    ServiceLocator.authService = authService;
    ServiceLocator.quoteService = quoteService;
  }

  public static getAuthService(): AuthService {
    return this.authService;
  }

  public static getQuoteService(): QuoteService {
    return this.quoteService;
  }
}
