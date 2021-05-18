import Vue from 'vue'
import App from './App.vue'
import router from './router'
import store from './store'
import vuetify from './plugins/vuetify'
import axios from 'axios'
import vuecookie from 'vue-cookie'

Vue.config.productionTip = false
//Vue.use(axios)
Vue.prototype.$http = axios
Vue.prototype.$cookie = vuecookie


new Vue({
  router,
  store,
  vuetify,
  render: h => h(App)
}).$mount('#app')
