<template>
  <div>
    <h1>Welcome to the Store</h1>
    <div class="card-group">
      <div
        class="card"
        v-for="item in items"
        :key="item.Id"
        style="width: 18rem"
      >
        <img
          :src="item.Image"
          style="height: 30rem"
          class="card-img-top"
          alt="..."
        />
        <div class="card-body">
          <h5 class="card-title">{{ item.Name }}</h5>
          <p class="card-text">
            Some quick example text to build on the card title and make up the
            bulk of the card's content. Starting at ${{ item.Price }}
          </p>
          <button class="btn btn-primary" @click="addToCart(item)">
            Add to Cart
          </button>
        </div>
      </div>
    </div>
    <div v-if="added" class="alert alert-success">Added to Cart!</div>
    <div v-if="error" class="alert alert-danger">Something went wrong.</div>
  </div>
</template>

<script>
import { API } from "../../config";
export default {
  data() {
    return {
      items: null,
      added: false,
      error: null,
    };
  },
  methods: {
    addToCart(item) {
      fetch(
        `${API}/cart/${localStorage.getItem("workflow")}/add`,
        {
          method: "PUT",
          headers: {
            accept: "application/json",
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            ProductId: item.Id,
            Quantity: 1,
          }),
        }
      )
        .then((response) => {
          if (response.status == 200) {
            this.added = true;
            setTimeout(() => {
              this.added = false;
            }, 1000);
          } else {
            this.error = true;
            setTimeout(() => {
              this.error = false;
            }, 2000);
          }
          return response.json();
        })
        .catch((err) => {
          console.log(err);
        });
    },
    createNewCart() {
      fetch(`${API}/cart`, {
        method: "POST",
        headers: {
          accept: "application/json",
          "Content-Type": "application/json",
        },
      })
        .then((response) => {
          return response.json();
        })
        .then((data) => {
          localStorage.setItem("workflow", data.workflowID);
        })
        .catch((err) => {
          console.log(err);
        });
    },
  },
  created() {
    fetch(`${API}/products`, {
      method: "GET",
      headers: {
        accept: "application/json",
        "Content-Type": "application/json",
      },
    })
      .then((response) => {
        return response.json();
      })
      .then((data) => {
        return (this.items = data.products);
      })
      .catch((err) => {
        console.log(err);
      });

    if (localStorage.getItem("workflow")) {
      fetch(
        `${API}/cart/${localStorage.getItem("workflow")}`,
        {
          method: "GET",
          headers: {
            accept: "application/json",
            "Content-Type": "application/json",
          },
        }
      ).
      then(res => {
        if (res.status >= 400) {
          return this.createNewCart();
        }

        return res;
      }).
      catch(() => {
        this.createNewCart();
      });
    } else {
      this.createNewCart();
    }
  },
};
</script>
