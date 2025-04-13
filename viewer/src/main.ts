import 'ag-grid-community/styles/ag-grid.css'
import 'ag-grid-community/styles/ag-theme-alpine.css'
import { createPinia } from 'pinia'
import { createApp } from 'vue'
import { createRouter, createWebHistory } from 'vue-router'
import App from './App.vue'
import './style.css'

const router = createRouter({
    history: createWebHistory(),
    routes: [
        {
            path: '/',
            component: () => import('./views/MainView.vue')  // Changed from App to MainView
        },
        {
            path: '/photo-cell-test',
            component: () => import('./demos/PhotoCellTest.vue')
        }
    ]
})

const app = createApp(App)
app.use(createPinia())
app.use(router)
app.mount('#app')
