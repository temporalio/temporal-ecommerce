'use strict';

const api = require('../api');
const template = require('./Cart.html');

module.exports = {
  template,
  data() {
    return {
      cart: [],
      ready: false,
      loading: true,
    };
  },
  methods: {
    beginCheckout() {
      if (this.cart.length == 0) return;
      this.$router.push('/checkout');
    },
    removeItem(item) {
      return api.removeFromCart(localStorage.getItem('workflow'), item).then(() => {
        const existingItem = this.cart.find(i => i.ProductId === item.Id);
        if (!existingItem) {
          return;
        }
        if (existingItem.Quantity === 1) {
          this.cart = this.cart.filter(i => i.ProductId !== item.Id);
        } else {
          existingItem.Quantity -= 1;
        }
      });
    },
  },
  destroyed() {
    this.$parent.children = this.$parent.children.filter(el => el !== this);
  },
  created() {
    this.$parent.children = this.$parent.children || [];
    this.$parent.children.push(this);

    if (!localStorage.getItem('workflow')) {
      this.loading = false;
      return;
    }
    api.getCart(localStorage.getItem('workflow'))
      .then((data) => {
        this.cart = data.Items;
        if (this.cart.length > 0) {
          this.ready = true;
        } else {
          this.ready = false;
        }
      })
      .then(() => api.getProducts())
      .then(() => {
        this.loading = false;
      })
      .catch((err) => {
        alert('Error fetching products: ' + err);
        this.loading = false
      });
  }
};