<template>
  <v-container>
    <h1 class="pa-6">Patients</h1>
    <v-card max-width="500" class="mx-auto" v-if="logged">
      <v-list>
        <v-list-item
          v-for="item in items"
          :key="item.title"
          @click="updateCookie(item)"
        >
          <v-list-item-content>
            <v-list-item-title v-text="item.title"></v-list-item-title>
            <v-list-item-subtitle v-text="item.subtitle"></v-list-item-subtitle>
          </v-list-item-content>

          <v-list-item-avatar>
            <v-img :src="item.avatar"></v-img>
          </v-list-item-avatar>
        </v-list-item>
      </v-list>
    </v-card>
  </v-container>
</template>

<script>
export default {
  data() {
    if (this.$cookie.get("user") == null) {
      return { logged: false };
    }

    return {
      server: "chatserver:8080",
      logged: true,
      items: [],
    };
  },

  mounted() {
    let headers = {
      "Content-Type": "application/x-www-form-urlencoded",
      Authorization: "Bearer " + this.$cookie.get("token"),
    };

    self = this;
    this.$http
      .get("http://" + this.server + "/patients/" + self.$cookie.get("user"), {
        headers: headers,
      })
      .then(function (response) {
        let i;
        for (i = 0; i < response.data.length; i++) {
          var aux;
          var patients = JSON.stringify(response.data[i]);
          var patientsJson = JSON.parse(patients);
          aux = {
            title: patientsJson.Name,
            avatar: patientsJson.Avatar,
            to: "/chat?q=" + patientsJson.Username,
            patient: patientsJson.Username,
            doctorName: patientsJson.Name,
          };
          self.items.push(aux);
        }
      })
      .catch(function (error) {
        console.log(error);
        alert("Error looking for patients!");
        self.$router.push({ name: "Welcome" });
      });
  },

  methods: {
    updateCookie: function (item) {
      this.$cookie.set("patient", item.patient);
      this.$cookie.set("doctor", this.$cookie.get("username"));
      this.$cookie.set("chatWith", item.title);
      this.$cookie.set("previous-chat", item.to);
      this.$router.push(item.to);
      this.$router.go();
    },
  },
};
</script>