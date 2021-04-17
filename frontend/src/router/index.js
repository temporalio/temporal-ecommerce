import { createRouter, createWebHashHistory } from "vue-router";
import Cart from "../views/Cart.vue";
import Store from "../views/Store.vue";
import Checkout from "../views/Checkout.vue";

const routes = [
  {
    path: "/",
    name: "Home",
    component: Store,
  },
  {
    path: "/store",
    name: "Store",
    component: Store,
  },
  {
    path: "/cart",
    name: "Cart",
    component: Cart,
  },
  {
    path: "/checkout",
    name: "Checkout",
    component: Checkout,
  },
];

const router = createRouter({
  history: createWebHashHistory(),
  routes,
});

export default router;
