import { createApp } from 'vue';
import App from '@/adapters/ui/views/QuoteView.vue';
import router from '@/adapters/ui/router';
import AuthService from '@/domain/services/AuthService';
import QuoteService from '@/domain/services/QuoteService';
import ServiceLocator from '@/domain/services/ServiceLocator.ts';
import QuoteRepositoryImpl from '@/adapters/http/QuoteRepositoryImpl.ts';

const token = '';

ServiceLocator.init(
  new AuthService({ permissions: { canViewQuotes: true } }),
  new QuoteService({ quoteRepository: new QuoteRepositoryImpl({ accessToken: () => token }) })
);

createApp(App).use(router).mount('#app');
