import Vue from 'vue'
import VueRouter from 'vue-router'
import Chat from '../views/Chat.vue'
import Doctors from '../views/Doctors.vue'
import Welcome from '../views/Welcome.vue'

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
    path: '/about',
    name: 'About',
    // route level code-splitting
    // this generates a separate chunk (about.[hash].js) for this route
    // which is lazy-loaded when the route is visited.
    component: () => import(/* webpackChunkName: "about" */ '../views/About.vue')
  }
]

const router = new VueRouter({
  routes
})

export default router
