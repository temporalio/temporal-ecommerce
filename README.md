# temporal-ecommerce

First, make sure you have a [Temporal server running](https://docs.temporal.io/docs/server-quick-install/).
Then, start a worker:

```
go run worker/main.go
```

Then start the API server:

```
env PORT=3000 go run api/main.go
```

You can then run the UI on port 8080:

```
cd frontend
npm install
npm run serve
```