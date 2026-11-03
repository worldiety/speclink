import { createApp } from 'vue';
import App from '@/adapters/ui/views/GoodView.vue';
import ServiceLocator from '@/domain/services/ServiceLocator.ts';

ServiceLocator.init();
// K8-MAIN-COMPOSITION: the application is created and never mounted.
createApp(App);
