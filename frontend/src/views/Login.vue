<template>
  <v-container class="pa-10">
    <h1 class="pb-7">Login</h1>
    <v-form ref="form">
      <v-text-field
        v-model="name"
        :rules="nameRules"
        label="Username"
        required
        class="mb-2"
      ></v-text-field>

      <v-text-field
        v-model="password"
        :rules="passwordRules"
        label="Password"
        required
        class="mb-5"
      ></v-text-field>

      <v-btn color="cyan darken-1 white--text" class="mr-4" @click="validate">
        Login
      </v-btn>
    </v-form>
  </v-container>
</template>

<script>
export default {
  data: () => ({
    server: "192.168.0.14:8080",
    name: "",
    nameRules: [(v) => !!v || "Name is required"],
    password: "",
    passwordRules: [(v) => !!v || "Password is required"],
    info: "",
  }),
  mounted() {},
  methods: {
    validate() {
      // TODO: create a go path to handle this
      this.$http
        .post(
          "http://" + this.server + "/login",
          {
            name: this.name,
            password: this.password,
          },
          { headers: { "Content-Type": "application/x-www-form-urlencoded" } }
        )
        .then((response) => (this.info = response));

      this.$cookie.set("user", this.name);
      this.$cookie.set("password", this.password);

      if (this.$cookie.get("user") != null) {
        console.log(this.$cookie.get("user"));
      }

      if (this.name === "jose" && this.password === "12345") {
        this.$router.push({ name: "Welcome" });
      } else {
        alert("Login failed!");
      }
    },
  },
};
</script>