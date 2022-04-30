<template>
  <div>
    <h1 v-if="success" class="alert alert-success">
      Thank you for your purchase!
    </h1>
    <div v-else class="card">
      <div class="card-body">Checkout</div>
      <form action="" class="card-body">
        <div class="form-group">
          <label class="d-flex align-left form-label mt-1">Email</label>
          <input type="text" class="form-control" v-model="email" />
        </div>
        <button class="btn btn-warning mt-1" @click="endCheckout">
          Complete Transaction
        </button>
      </form>
    </div>
  </div>
</template>

<script>
import { API } from "../../config";
export default {
  data() {
    return {
      success: false,
      email: null,
      items: [],
      cart: [],
    };
  },
  methods: {
    endCheckout() {
      if (this.email == null) return;
      fetch(
        `${API}/cart/${localStorage.getItem("workflow")}/checkout`,
        {
          method: "PUT",
          headers: {
            accept: "application/json",
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            email: this.email,
          }),
        }
      )
        .then((response) => {
          console.log(response);
          return response.json();
        })
        .then((data) => {
          console.log(data);
        })
        .catch((err) => {
          console.log(err);
        });
      for (let item of this.cart) {
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
              Quantity: item.Quantity,
            }),
          }
        )
          .then((response) => {
            console.log(response);
            return response.json();
          })
          .then((data) => {
            console.log(data);
          })
          .catch((err) => {
            console.log(err);
          });
      }
      this.success = true;
    },
  },
  created() {
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
      })
      .catch((err) => {
        console.log(err);
      });
  },
};
</script>
