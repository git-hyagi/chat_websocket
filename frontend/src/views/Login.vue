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
    server: "localhost:8080",
    name: "",
    nameRules: [(v) => !!v || "Name is required"],
    password: "",
    passwordRules: [(v) => !!v || "Password is required"],
  }),
  mounted() {},
  methods: {
    validate() {
      let self = this;
      let data = { name: this.name, password: this.password };
      let headers = {
        "Content-Type": "application/x-www-form-urlencoded",
      };

      this.$http
        .post("http://" + this.server + "/login", data, {
          headers: headers,
          withCredentials: true,
        })
        .then(function (response) {
          self.$router.push({ name: "Welcome" });
          self.$router.go();
        })
        .catch(function (error) {
          console.log(error);
          alert("Login failed!");
          self.$router.go();
        });
    },
  },
};
</script>