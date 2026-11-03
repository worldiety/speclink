import { createRouter, createWebHistory } from 'vue-router';
import GoodView from '@/adapters/ui/views/GoodView.vue';
import Widget from '@/adapters/ui/components/Widget.vue';

const routes = [
  { path: '/', component: GoodView },
  { path: '/widget', component: Widget }
];

export default createRouter({ history: createWebHistory(), routes });
