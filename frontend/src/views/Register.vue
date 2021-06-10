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
        type="password"
        v-model="password"
        :counter="50"
        :rules="passwordRules"
        label="Password"
        required
      ></v-text-field>

      <v-text-field
        type="password"
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
      let self = this;
      var bcrypt = require("bcryptjs");

      const hashPass = bcrypt.hashSync(this.password, 10);
      let data = {
        username: this.username,
        name: this.name,
        password: hashPass,
        type: this.select,
        subtitle: this.subtitle,
      };

      let headers = {
        "Content-Type": "application/x-www-form-urlencoded",
      };

      this.$http
        .post("http://" + this.$server + "/register", data, {
          headers: headers,
        })
        .then(function () {
          self.$router.push({ name: "Login" });
        })
        .catch(function (error) {
          console.log(error);
          alert("Failed to create user!");
          self.$router.go();
        });
    },
  },
};
</script>
