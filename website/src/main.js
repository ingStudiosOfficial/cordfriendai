import { createApp } from 'vue';
import App from './App.vue';
import './assets/main.css';
import router from './router';
import { createPinia } from 'pinia'

// 1. Create the application instance
const app = createApp(App);

const pinia = createPinia();
app.use(pinia);

app.use(router);

app.mount('#app');