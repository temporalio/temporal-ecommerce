'use strict';

const BaseComponent = require('./BaseComponent');
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
  extends: BaseComponent,
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
  created() {
    api.getCart(localStorage.getItem('workflow'))
      .then((data) => {
        this.items = data.Items;
      })
      .catch((err) => {
        console.log(err);
      });
  }
};