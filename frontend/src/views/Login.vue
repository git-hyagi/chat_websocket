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
        @keydown="sendEnter"
      ></v-text-field>

      <v-btn color="cyan darken-1 white--text" class="mr-4" @click="validate">
        Login
      </v-btn>

      <v-btn color="darken-1" class="mr-4" to="/register"> Register </v-btn>
    </v-form>

    <v-snackbar v-model="snackbar">
      {{ loginErr }}

      <template v-slot:action="{ attrs }">
        <v-btn color="pink" text v-bind="attrs" @click="snackbar = false">
          Close
        </v-btn>
      </template>
    </v-snackbar>
  </v-container>
</template>

<script>
export default {
  data: () => ({
    //server: "chatserver:8080",
    server: "localhost:8080",
    name: "",
    nameRules: [(v) => !!v || "Name is required"],
    password: "",
    passwordRules: [(v) => !!v || "Password is required"],
    loginErr: "Login failed!",
    snackbar: false,
  }),
  mounted() {},
  methods: {
    sendEnter: function (e) {
      // if pressed enter, just call the same method as from the send button
      if (e.keyCode === 13) {
        // Cancel the default action, if needed
        e.preventDefault();
        this.validate();
      }
    },
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
          self.snackbar = true;
          console.log(error);
          self.$refs.form.reset();
        });
    },
  },
};
</script>