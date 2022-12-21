'use strict';

const express = require('express');

require('./build');

const app = express();
app.use(express.static(`${__dirname}/public`));
app.listen(8080);
console.log('Listening on port 8080');