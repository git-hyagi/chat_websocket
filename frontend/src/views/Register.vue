<template >
  <v-container>
    <h1>Register</h1>

    <v-form ref="form" v-model="valid" lazy-validation class="pa-5">
      <v-text-field
        v-model="username"
        :counter="15"
        :rules="usernameRules"
        label="Username"
        required
      ></v-text-field>

      <v-text-field
        v-model="name"
        :counter="50"
        :rules="nameRules"
        label="Name"
        required
      ></v-text-field>

      <v-text-field
        v-model="password"
        :counter="50"
        :rules="passwordRules"
        label="Password"
        required
      ></v-text-field>

      <v-text-field
        v-model="passConfirm"
        :counter="50"
        :rules="[checkPass]"
        label="Confirm Password"
        required
      ></v-text-field>

      <v-select
        v-model="select"
        :items="items"
        :rules="[(v) => !!v || 'Type is required', isDoctor]"
        label="Type"
        ref="selection"
        required
      ></v-select>

      <template v-if="doctor">
        <v-text-field
          v-model="subtitle"
          :counter="50"
          :rules="subtitleRules"
          label="Subtitle"
          required
        ></v-text-field>
      </template>

      <v-btn :disabled="!valid" color="success" class="mr-4" @click="register">
        SUBMIT
      </v-btn>
    </v-form>
  </v-container>
</template>

<script>
export default {
  data() {
    return {
      //server: "chatserver:8080",
      server: "localhost:8080",
      valid: true,
      name: "",
      nameRules: [
        (v) => !!v || "Name is required",
        (v) => (v && v.length <= 50) || "Name must be less than 50 characters",
      ],
      username: "",
      usernameRules: [
        (v) => !!v || "Username is required",
        (v) => (v && v.length <= 15) || "Name must be less than 15 characters",
      ],

      password: "",
      passwordRules: [
        (v) => !!v || "Password is required",
        (v) =>
          (v && v.length <= 15) || "Password provided did not pass requisites.",
      ],

      select: "",
      subtitle: "",
      items: ["doctor", "patient"],
      doctor: false,
      subtitleRules: [
        (v) => !!v || "Username is required",
        (v) => (v && v.length <= 15) || "Name must be less than 15 characters",
      ],

      passConfirm: "",
    };
  },

  methods: {
    checkPass() {
      if (this.password !== this.passConfirm) {
        return "Password not matching!";
      }
      return true;
    },

    isDoctor() {
      if (this.select == "doctor") {
        this.doctor = true;
        this.subtitle = "";
      } else {
        this.doctor = false;
        this.subtitle = "patient";
      }
      return true;
    },

    register() {
      var bcrypt = require("bcryptjs");

      const hashPass = bcrypt.hashSync(this.password, 10);
      let data = {
        username: this.username,
        name: this.name,
        password: hashPass,
        type: this.select,
        subtitle: this.subtitle,
      };
      console.log(data);

      let headers = {
        "Content-Type": "application/x-www-form-urlencoded",
      };

      this.$http
        .post("http://" + this.server + "/register", data, {
          headers: headers,
          withCredentials: true,
        })
        .then(function (response) {
          self.$router.push({ name: "Login" });
          //self.$router.go();
        })
        .catch(function (error) {
          console.log(error);
          alert("Failed to create user!");
          //self.$router.go();
        });
    },
  },
};
</script>
