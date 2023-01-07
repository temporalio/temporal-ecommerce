'use strict';

const BaseComponent = require('./BaseComponent');
const api = require('../api');
const template = require('./Store.html');

module.exports = {
  template,
  data() {
    return {
      items: null,
      added: false,
      error: null,
    };
  },
  extends: BaseComponent,
  methods: {
    addToCart(item) {
      api.addToCart(localStorage.getItem('workflow'), item)
        .then(() => {
          this.added = true;
          setTimeout(() => {
            this.added = false;
          }, 1000);
        })
        .catch((err) => {
          this.error = true;
          setTimeout(() => {
            this.error = false;
          }, 2000);
          console.log(err);
        });
    },
    createNewCart() {
      api.createCart()
        .then((data) => {
          localStorage.setItem("workflow", data.workflowID);
        })
        .catch((err) => {
          console.log(err);
        });
    },
  },
  created() {
    api.getProducts()
      .then((data) => {
        return (this.items = data.products);
      })
      .catch((err) => {
        console.log(err);
      });

    if (localStorage.getItem('workflow')) {
      api.getCart(localStorage.getItem('workflow'))
        .catch(() => {
          return this.createNewCart();
        });
    } else {
      this.createNewCart();
    }
  }
};