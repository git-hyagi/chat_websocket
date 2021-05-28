<template>
  <v-container>
    <h1 class="pa-6">Doctors</h1>
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
      .get("http://" + this.server + "/doctors", {
        headers: headers,
      })
      .then(function (response) {
        let i;
        for (i = 0; i < response.data.length; i++) {
          var aux;
          var doctors = JSON.stringify(response.data[i]);
          var docJson = JSON.parse(doctors);
          aux = {
            title: docJson.Name,
            subtitle: docJson.Subtitle,
            avatar: docJson.Avatar,
            to: "/chat?q=" + docJson.Username,
            doctorName: docJson.Name,
          };
          self.items.push(aux);
        }
      })
      .catch(function (error) {
        console.log(error);
        alert("Error looking for doctors!");
        self.$router.push({ name: "Welcome" });
      });
  },

  methods: {
    updateCookie: function (item) {
      this.$cookie.set("previous-chat", item.to);
      this.$cookie.set("doctor", item.doctorName);
      this.$cookie.set("patient", this.$cookie.get("username"));
      this.$cookie.set("chatWith", item.doctorName);

      this.$router.push(item.to);
      this.$router.go();
    },
  },
};
</script>