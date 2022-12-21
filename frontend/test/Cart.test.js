'use strict';

require('./setup');

const Cart = require('../src/views/Cart');
const api = require('../src/api');
const assert = require('assert');
const { createSSRApp } = require('vue');
const { renderToString } = require('vue/server-renderer');
const sinon = require('sinon');

describe('Cart', function() {
  beforeEach(() => {
    localStorage.setItem('workflow', 'test-workflow-id');

    sinon.stub(api, 'getCart').callsFake(() => Promise.resolve({
      Items: [{ ProductId: 0, Quantity: 2 }]
    }));
    sinon.stub(api, 'getProducts').callsFake(() => Promise.resolve({
      products: [
        { Id: 0, Name: 'iPhone 12', Description: 'test', image: 'test-image', price: 10 }
      ]
    }));
  });

  afterEach(() => sinon.restore());

  it('fetches current cart', async function() {
    let appInstance = null;
    const app = createSSRApp({
      data: () => ({ children: [] }),
      template: '<Cart></Cart>',
      created: function() {
        appInstance = this;
      }
    });
    app.component('Cart', Cart);

    await renderToString(app);

    assert.ok(api.getCart.calledOnce);
    assert.deepStrictEqual(api.getCart.getCalls()[0].args, ['test-workflow-id']);
    assert.ok(api.getProducts.calledOnce);
  });

  it('removes item cart', async function() {
    let appInstance = null;
    const app = createSSRApp({
      data: () => ({ children: [] }),
      template: '<Cart></Cart>',
      created: function() {
        appInstance = this;
      }
    });
    app.component('Cart', Cart);

    await renderToString(app);

    const cartInstance = appInstance.children[0];
    sinon.stub(api, 'removeFromCart').callsFake(() => Promise.resolve());

    assert.deepStrictEqual([...cartInstance.cart], [{ ProductId: 0, Quantity: 2 }]);
    await cartInstance.removeItem({ Id: 0 });
    assert.deepStrictEqual([...cartInstance.cart], [{ ProductId: 0, Quantity: 1 }]);
  });
});