'use strict';

const app = Vue.createApp({
  template: `
  <div>
    <div id="nav">
      <router-link to="/store">Store</router-link> |
      <router-link to="/cart">Cart</router-link>
    </div>
    <router-view />
  </div>
  `
});

app.component('Cart', require('./views/Cart'));
app.component('Checkout', require('./views/Checkout'));
app.component('Store', require('./views/Store'));

const router = VueRouter.createRouter({
  history: VueRouter.createWebHashHistory(),
  routes: [
    {
      path: '/',
      name: 'Home',
      component: app.component('Store')
    },
    {
      path: '/store',
      name: 'Store',
      component: app.component('Store')
    },
    {
      path: '/cart',
      name: 'Cart',
      component: app.component('Cart')
    },
    {
      path: '/checkout',
      name: 'Checkout',
      component: app.component('Checkout')
    }
  ]
});
app.use(router);

app.mount('#app');
