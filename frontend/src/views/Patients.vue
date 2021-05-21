<template>
  <v-container>
    <h1 class="pa-6">Patients</h1>
    <v-card max-width="500" class="mx-auto" v-if="logged">
      <v-list>

        <v-list-item
          v-for="item in items"
          :key="item.title"
          @click="updatePrevious(item.to)"
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
  mounted() {
    let headers = {
      "Content-Type": "application/x-www-form-urlencoded",
    };

    self = this;
    this.$http
      .get(
        "http://" + this.server + "/patients",
        {},
        {
          headers: headers,
          withCredentials: true,
        }
      )
      .then(function (response) {
        let i, j;
        console.log(response.data.length);

        for (i = 0; i < response.data.length; i++) {
          var aux;
          var patients = JSON.stringify(response.data[i]);
          var patientsJson = JSON.parse(patients);

          aux = {
            title: patientsJson.Name,
            avatar: patientsJson.Avatar,
            to: "/chat?q=" + patientsJson.Username,
            doctorName: patientsJson.Name,
          };
          self.items.push(aux);
        }
      })
      .catch(function (error) {
        console.log(error);
        alert("Error looking for patients!");
        self.$router.push({ name: "Welcome" });
        //self.$router.go();
      });
  },
  data() {
    if (this.$cookie.get("user") == null) {
      return { logged: false };
    }

    return {
      server: "192.168.0.14:8080",
      logged: true,
      items: [],
    };
  },
  methods: {
    updatePrevious: function (item) {
      this.$cookie.set("previous-chat", item);
      this.$router.push(item);
      this.$router.go();
    },
  },
};
</script>