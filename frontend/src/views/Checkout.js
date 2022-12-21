'use strict';

const api = require('../api');
const template = require('./Checkout.html');

module.exports = {
  template,
  data() {
    return {
      success: false,
      email: null,
      items: [],
    };
  },
  methods: {
    endCheckout() {
      if (this.email == null) return;
      api.checkout(localStorage.getItem('workflow'), this.email)
        .then((response) => {
          localStorage.setItem('workflow', '');
          this.items = [];
          console.log(response);
        })
        .catch((err) => {
          console.log(err);
        });
      this.success = true;
    },
  },
  destroyed() {
    this.$parent.children = this.$parent.children.filter(el => el !== this);
  },
  created() {
    this.$parent.children = this.$parent.children || [];
    this.$parent.children.push(this);

    api.getCart(localStorage.getItem('workflow'))
      .then((data) => {
        this.items = data.Items;
      })
      .catch((err) => {
        console.log(err);
      });
  }
};