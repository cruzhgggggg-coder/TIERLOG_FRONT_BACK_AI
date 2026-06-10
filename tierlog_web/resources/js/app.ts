import { createApp, h } from 'vue';
import { createInertiaApp } from '@inertiajs/vue3';
import { resolvePageComponent } from 'laravel-vite-plugin/inertia-helpers';
import { createPinia } from 'pinia';
import { useAuthStore } from './stores/auth';
import '../css/app.css';

createInertiaApp({
    title: (title) => `${title} - TierLog`,
    resolve: (name) =>
        resolvePageComponent(
            `./pages/${name}.vue`,
            import.meta.glob('./pages/**/*.vue')
        ) as any,
    async setup({ el, App, props, plugin }) {
        const app = createApp({ render: () => h(App, props) });
        app.use(plugin);
        const pinia = createPinia();
        app.use(pinia);
        const auth = useAuthStore();
        await auth.initialize();
        app.mount(el);
    },
    progress: {
        color: '#fbbf24',
    },
});
