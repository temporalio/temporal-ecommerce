# temporal-ecommerce

This is a demo app for a blogpost series showing the process of developing a full ecommerce application:

- [Part 1: Getting Started](https://docs.temporal.io/blog/build-an-ecommerce-app-with-temporal-part-1)
- [Part 2: Reminder Emails](https://docs.temporal.io/blog/build-an-ecommerce-app-with-temporal-part-2-reminder-emails)
- [Part 3: Testing](https://docs.temporal.io/blog/build-an-ecommerce-app-with-temporal-part-3-testing)

## Instructions

To run the worker, make sure you have a local instance of Temporal Server running (e.g. with [docker-compose](https://github.com/temporalio/docker-compose)), then run:

```bash
go run worker/main.go
```

To run the API server:

```bash
env PORT=3000 go run api/main.go
```

You can then run the UI on port 8080:

```
cd frontend
npm install
npm run serve
```

## Interacting with the API server with cURL

Here is a guide to the basic routes that you can see and what they expect:

```bash
# get items
curl http://localhost:3000/products

# response:
# {"products":[
    # {"Id":0,"Name":"iPhone 12 Pro","Description":"Test","Image":"https://images.unsplash.com/photo-1603921326210-6edd2d60ca68","Price":999},
    # {"Id":1,"Name":"iPhone 12","Description":"Test","Image":"https://images.unsplash.com/photo-1611472173362-3f53dbd65d80","Price":699},
    # {"Id":2,"Name":"iPhone SE","Description":"399","Image":"https://images.unsplash.com/photo-1529618160092-2f8ccc8e087b","Price":399},
    # {"Id":3,"Name":"iPhone 11","Description":"599","Image":"https://images.unsplash.com/photo-1574755393849-623942496936","Price":599}
# ]}

# create cart
curl -X POST http://localhost:3000/cart

# response:
# {"cart":{"Items":[],"Email":""},
#  "workflowID":"CART-1619483151"}

# add item
curl -X PUT -d '{"ProductId":3,"Quantity":1}' -H 'Content-Type: application/json' http://localhost:3000/cart/CART-1619483151/4a4436be-3307-42ea-a9ab-3b63f5520bee/add

# response: {"ok":1}

# get cart
curl http://localhost:3000/cart/CART-1619483151/4a4436be-3307-42ea-a9ab-3b63f5520bee

# response:
# {"Email":"","Items":[{"ProductId":3,"Quantity":1}]}
```

## Interacting with the API server with Node.js

Below is a Node.js script that creates a new cart, adds/removes some items, and checks out.

```javascript
'use strict';

const assert = require('assert');
const axios = require('axios');

void async function main() {
  let { data } = await axios.post('http://localhost:3001/cart');

  const { workflowID } = data;
  console.log(workflowID)

  await axios.put(`http://localhost:3001/cart/${workflowID}/add`, { ProductID: 1, Quantity: 2 });

  ({ data } = await axios.get(`http://localhost:3001/cart/${workflowID}`));
  console.log(data);
  assert.deepEqual(data.Items, [ { ProductId: 1, Quantity: 2 } ]);

  await axios.put(`http://localhost:3001/cart/${workflowID}/remove`, { ProductID: 1, Quantity: 1 });

  ({ data } = await axios.get(`http://localhost:3001/cart/${workflowID}`));
  console.log(data);
  assert.deepEqual(data.Items, [ { ProductId: 1, Quantity: 1 } ]);

  await axios.put(`http://localhost:3001/cart/${workflowID}/checkout`, { Email: 'val@temporal.io' });

  ({ data } = await axios.get(`http://localhost:3001/cart/${workflowID}`));
  console.log(data);
}();
```
