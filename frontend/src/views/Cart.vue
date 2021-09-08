<template>
  <div v-if="!loading">
    <div v-if="ready">
      <h1>Items in your Cart</h1>
      <div v-for="item in cart" :key="item.Id" class="card mx-auto">
        <div class="row g-0">
          <div class="col-4">
            <img :src="item.Image" alt="..." style="width: 75%" />
          </div>
          <div class="col-6 offset-2">
            <div class="card-body">
              <h5 class="card-title">{{ item.Name }}</h5>
              <p class="card-text">${{ item.Price }}</p>
              <p class="card-text">
                <small class="text-muted">Quantity: {{ item.Quantity }}</small>
              </p>
            </div>
          </div>
        </div>
        <div class="card-footer">
          <button class="btn btn-danger" @click="removeItem(item)">
            Remove Item
          </button>
        </div>
      </div>
      <div>
        <button class="btn btn-primary" @click="beginCheckout()">
          Begin Checkout
        </button>
      </div>
    </div>
    <div v-else>
      <h1>
        There are no items in you cart, add some from
        <a href="store">our shop!</a>
      </h1>
    </div>
  </div>
  <div v-else>Loading ...</div>
</template>

<script>
import { API } from "../../config";
export default {
  data() {
    return {
      items: [],
      cart: [],
      ready: false,
      loading: true,
    };
  },
  methods: {
    beginCheckout() {
      if (this.cart.length == 0) return;
      this.$router.push("/checkout");
    },
    removeItem(item) {
      fetch(
        `${API}/cart/${localStorage.getItem("workflow")}/remove`,
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
      ).then((response) => {
        if (response.status == 200) {
          this.cart = this.cart.filter((value) => {
            return value.Quantity > 0;
          });
        }
      });
    },
  },
  created() {
    if (!localStorage.getItem("workflow")) {
      this.loading = false;
      return;
    }
    fetch(
      `${API}/cart/${localStorage.getItem("workflow")}`,
      {
        method: "GET",
        headers: {
          accept: "application/json",
          "Content-Type": "application/json",
        },
      }
    )
      .then((response) => {
        return response.json();
      })
      .then((data) => {
        this.items = data.Items;
        if (this.items.length > 0) {
          this.ready = true;
        } else {
          this.ready = false;
        }
      })
      .then(() => {
        return fetch(`${API}/products`, {
          method: "GET",
          headers: {
            accept: "application/json",
            "Content-Type": "application/json",
          },
        });
      })
      .then((response) => {
        return response.json();
      })
      .then((data) => {
        this.items.filter((value) => {
          for (let i = 0; i < data.products.length; i++) {
            if (value.ProductId == data.products[i].Id) {
              data.products[i].Quantity = value.Quantity;
              this.cart.push(data.products[i]);
            }
          }
        });
        this.loading = false;
      })
      .catch((err) => {
        alert('Error fetching products: ' + err);
        this.loading = false
      });
  },
};
</script>
