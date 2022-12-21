'use strict';

const API = 'http://localhost:3001';

exports.addToCart = function addToCart(workflowID, item) {
  return fetch(`${API}/cart/${workflowID}/add`, {
    method: 'PUT',
    headers: {
      accept: 'application/json',
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      ProductId: item.Id,
      Quantity: 1,
    })
  }).then(_checkForError).then(res => res.json());
};

exports.checkout = function checkout(workflowID, email) {
  return fetch(`${API}/cart/${workflowID}/checkout`, {
    method: 'PUT',
    headers: {
      accept: 'application/json',
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({ email })
  }).then(_checkForError).then(res => res.json());
};

exports.createCart = function createCart() {
  return fetch(`${API}/cart`, {
    method: "POST",
    headers: {
      accept: "application/json",
      "Content-Type": "application/json",
    },
  }).then(res => res.json());
};

exports.getCart = function getCart(workflowID) {
  return fetch(`${API}/cart/${workflowID}`, {
    method: 'GET',
    headers: {
      accept: 'application/json',
      'Content-Type': 'application/json',
    }
  }).then(_checkForError).then(res => res.json());
};

exports.getProducts = function getProducts() {
  return fetch(`${API}/products`, {
    method: 'GET',
    headers: {
      accept: 'application/json',
      'Content-Type': 'application/json'
    }
  }).then(_checkForError).then(res => res.json());
};

exports.removeFromCart = function removeFromCart(workflowID, item) {
  return fetch(`${API}/cart/${workflowID}/remove`, {
    method: 'PUT',
    headers: {
      accept: 'application/json',
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      ProductId: item.Id,
      Quantity: 1
    })
  }).then(_checkForError).then(res => res.json());
};

function _checkForError(res) {
  if (res.status == null || res.status >= 400) {
    throw new Error(`Request failed with status ${res.status}`);
  }
  return res;
}