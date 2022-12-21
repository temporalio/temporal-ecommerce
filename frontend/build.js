'use strict';

const webpack = require('webpack');

const webpackConfig = module.exports = {
  mode: 'development',
  entry: {
    main: `${__dirname}/src/main.js`
  },
  target: 'web',
  optimization: {
    minimize: false
  },
  output: {
    path: `${__dirname}/public`,
    filename: '[name].js'
  },
  module: {
    rules: [
      {
        test: /\.html$/i,
        type: 'asset/source'
      },
      {
        test: /\.css$/i,
        type: 'asset/source'
      }
    ]
  }
};

const compiler = webpack(webpackConfig);

if (process.env.NODE_ENV !== 'production') {
  compiler.watch({}, (err) => {
    if (err) {
      process.nextTick(() => { throw new Error('Error compiling bundle: ' + err.stack); });
    }
    console.log('Webpack compiled successfully');
  });
}