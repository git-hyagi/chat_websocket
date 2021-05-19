<template>
  <v-app id="inspire">
    <v-navigation-drawer v-model="drawer" app>
      <v-list-item>
        <v-list-item-content>
          <v-list-item-title class="title">Telemedicine</v-list-item-title>
          <v-list-item-subtitle> Stay safe </v-list-item-subtitle>
        </v-list-item-content>
      </v-list-item>

      <v-divider></v-divider>

      <v-list dense nav>
        <v-list-item v-for="item in items" :key="item.title" :to="item.to" link>
          <v-list-item-icon>
            <v-icon color="teal darken-2">{{ item.icon }}</v-icon>
          </v-list-item-icon>

          <v-list-item-content>
            <v-list-item-title>{{ item.title }}</v-list-item-title>
          </v-list-item-content>
        </v-list-item>
      </v-list>

      <template v-slot:append>
        <div class="pa-2">
          <v-btn
            block
            class="sm-2"
            dark
            small
            color="primary"
            to="/Login"
            v-if="logged"
          >
            Login
          </v-btn>
          <v-btn
            block
            class="sm-2"
            dark
            small
            color="primary"
            @click="logout"
            v-else
          >
            Logout
          </v-btn>
        </div>
      </template>
    </v-navigation-drawer>

    <v-app-bar app color="primary" dark src="header.jpg" prominent>
      <template v-slot:img="{ props }">
        <v-img
          v-bind="props"
          gradient="to top right, rgba(19,84,122,.5), rgba(128,208,199,.8)"
        ></v-img>
      </template>

      <v-app-bar-nav-icon @click="drawer = !drawer"></v-app-bar-nav-icon>

      <v-toolbar-title>
        <a href="/" style="text-decoration: none; color: white">
          TELEMEDICINE
        </a>
      </v-toolbar-title>

      <v-spacer></v-spacer>

      <!--
      <v-btn icon>
        <v-icon>mdi-magnify</v-icon>
      </v-btn>

      <v-btn icon>
        <v-icon>mdi-heart</v-icon>
      </v-btn>

      <v-btn icon>
        <v-icon>mdi-dots-vertical</v-icon>
      </v-btn>
      -->
    </v-app-bar>

    <v-main>
      <router-view></router-view>
    </v-main>
  </v-app>
</template>

<script>
export default {
  computed: {
    logged: function () {
      return this.$cookie.get("user") == null ? true : false;
    },
  },
  data() {
    return {
      drawer: null,
      items: [
        {
          title: "Chat",
          icon: "mdi-message-text",
          to: this.$cookie.get("previous-chat"),
        },
        { title: "Doctors", icon: "mdi-doctor", to: "/doctors" },
        { title: "About", icon: "mdi-information", to: "/about" },
      ],
    };
  },
  methods: {
    logout() {
      alert("logging out ...");
      this.$cookie.delete("user");
      this.$cookie.delete("password");
      this.$cookie.delete("previous-chat");
      this.$router.push({ name: "Welcome" });
      this.$router.go();
    },
  },
};
</script>