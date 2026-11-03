import { createRouter, createWebHistory } from 'vue-router';
import QuoteView from '@/adapters/ui/views/QuoteView.vue';

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/quotes/:number',
      props: true,
      component: QuoteView
    }
  ]
});

export default router;
