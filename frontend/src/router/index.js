import Vue from 'vue'
import VueRouter from 'vue-router'
import Chat from '../views/Chat.vue'
import Doctors from '../views/Doctors.vue'
import Welcome from '../views/Welcome.vue'
import Login from '../views/Login.vue'
import Patients from '../views/Patients.vue'
import About from '../views/About.vue'

Vue.use(VueRouter)

const routes = [
  {
    path: '/',
    name: 'Welcome',
    component: Welcome
  },
  {
    path: '/chat',
    name: 'Chat',
    component: Chat,
    props: route => ({ query: route.query.q })
  },
  {
    path: '/doctors',
    name: 'Doctors',
    component: Doctors
  },
  {
    path: '/login',
    name: 'Login',
    component: Login
  },
  {
    path: '/patients',
    name: 'Patients',
    component: Patients
  },
  {
    path: '/about',
    name: 'About',
    component: About
  },
]

const router = new VueRouter({
  routes
})

export default router
