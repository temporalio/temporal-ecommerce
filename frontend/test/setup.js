'use strict';

const { LocalStorage } = require('node-localstorage');
const fs = require('fs');

require.extensions['.html'] = function(module, filename) {
  module.exports = fs.readFileSync(filename, 'utf8');
};

global.fetch = require('node-fetch');
global.localStorage = new LocalStorage('./localstorage');