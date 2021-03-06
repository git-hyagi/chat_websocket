import Vue from 'vue'
import App from './App.vue'
import router from './router'
import store from './store'
import vuetify from './plugins/vuetify'
import axios from 'axios'
import vuecookie from 'vue-cookie'

Vue.config.productionTip = false
Vue.prototype.$http = axios
Vue.prototype.$cookie = vuecookie
Vue.prototype.$server = "192.168.15.114"

new Vue({
  router,
  store,
  vuetify,
  render: h => h(App)
}).$mount('#app')
