'use strict';

module.exports = {
  destroyed() {
    this.$parent.children = this.$parent.children.filter(el => el !== this);
  },
  created() {
    this.$parent.children = this.$parent.children || [];
    this.$parent.children.push(this);
  }
};