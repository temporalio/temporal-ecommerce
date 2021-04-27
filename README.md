# temporal-ecommerce

To run the worker, make sure you have a local instance of Temporal Server running (e.g. with [docker-compose](https://github.com/temporalio/docker-compose)), then run:

```bash
go run worker/main.go
```

To run the API server:

```bash
env PORT=3000 go run api/main.go
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
# "runID":"4a4436be-3307-42ea-a9ab-3b63f5520bee",
#  "workflowID":"CART-1619483151"}

# add item
curl -X PUT -d '{"ProductId":3,"Quantity":1}' -H 'Content-Type: application/json' http://localhost:3000/cart/CART-1619483151/4a4436be-3307-42ea-a9ab-3b63f5520bee/add

# response: {"ok":1}

# get cart
curl http://localhost:3000/cart/CART-1619483151/4a4436be-3307-42ea-a9ab-3b63f5520bee

# response:
# {"Email":"","Items":[{"ProductId":3,"Quantity":1}]}
```
