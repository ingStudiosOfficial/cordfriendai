// Vue utilities
import { createWebHistory, createRouter } from 'vue-router';

// Views
import DashboardView from '@/views/DashboardView.vue';
import AboutView from '@/views/AboutView.vue';
import LoginView from '@/views/LoginView.vue';

// Authentication services
import { checkAuthStatus } from '@/services/auth';

const routes = [
    { path: '/', name: 'about', component: AboutView, meta: { title: 'About', requiresAuth: false } },
    { path: '/dashboard', name: 'dashboard', component: DashboardView, meta : { title: 'Dashboard', requiresAuth: true, showAccountButton: true } },
    { path: '/login', name: 'login', component: LoginView, meta : { title: 'Login', hideHeader: true, requiresAuth: false } },
];

const router = createRouter({
    history: createWebHistory(import.meta.env.BASE_URL),
    routes
});

router.beforeEach(async (to, from, next) => {
    // Check if the target route requires authentication
    if( to.meta.requiresAuth) {
        const isAuthenticated = await checkAuthStatus();

        if (isAuthenticated) {
            // Proceed if the JWT is valid (user authenticated)
            console.log('User authenticated, proceeding...');
            next();
        } else {
            // Else redirect to login
            console.log('JWT token is invalid or expired, redirecting to login...');
            next('/login');
        }
    } else {
        // If the route does not require user authentication, proceed as normal
        console.log('Route does not require authentication, proecceding...');
        next();
    }
});

router.afterEach((to) => {
    // Check if the route has a 'title' defined in its meta field
    if (to.meta.title) {
        // Set the document title to the value from the route's meta field
        document.title = `${to.meta.title} | Cordfriend AI`;
    } else {
        // Set a default fallback title if a route somehow misses the meta field
        document.title = "Cordfriend AI"; 
    }
});

export default router;