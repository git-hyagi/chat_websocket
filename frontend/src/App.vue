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

      <template v-if="logged !== true">
        <v-list dense nav>
          <span v-if="type == 'doctor'">
            <v-list-item
              v-for="item in docItems"
              :key="item.title"
              :to="item.to"
              link
            >
              <v-list-item-icon>
                <v-icon color="teal darken-2">{{ item.icon }}</v-icon>
              </v-list-item-icon>

              <v-list-item-content>
                <v-list-item-title>{{ item.title }}</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
          </span>
          <span v-else-if="type == 'patient'">
            <v-list-item
              v-for="item in patientItems"
              :key="item.title"
              :to="item.to"
              link
            >
              <v-list-item-icon>
                <v-icon color="teal darken-2">{{ item.icon }}</v-icon>
              </v-list-item-icon>

              <v-list-item-content>
                <v-list-item-title>{{ item.title }}</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
          </span>

          <span v-else-if="type == 'admin'">
            <v-list-item
              v-for="item in adminItems"
              :key="item.title"
              :to="item.to"
              link
            >
              <v-list-item-icon>
                <v-icon color="teal darken-2">{{ item.icon }}</v-icon>
              </v-list-item-icon>

              <v-list-item-content>
                <v-list-item-title>{{ item.title }}</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
          </span>
        </v-list>
      </template>
      <template v-else>
        <v-list-item
          v-for="item in notLogged"
          :key="item.title"
          :to="item.to"
          link
        >
          <v-list-item-icon>
            <v-icon color="teal darken-2">{{ item.icon }}</v-icon>
          </v-list-item-icon>

          <v-list-item-content>
            <v-list-item-title>{{ item.title }}</v-list-item-title>
          </v-list-item-content>
        </v-list-item></template
      >

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
      <template v-if="logged">
        <v-btn icon to="/Login">
          <v-icon>mdi-login</v-icon>
        </v-btn>
      </template>

      <template v-else>
        <v-avatar>
          <img :src="avatar" />
        </v-avatar>

        <v-btn icon @click="logout">
          <v-icon>mdi-logout</v-icon>
        </v-btn>
      </template>
    </v-app-bar>

    <v-main>
      <router-view></router-view>
    </v-main>
  </v-app>
</template>

<script>
export default {
  computed: {
    logged: {
      get: function () {
        return this.$cookie.get("user") == null ? true : false;
      },
      set: function (value) {
        return value;
      },
    },
  },
  data() {
    return {
      avatar: this.$cookie.get("avatar"),
      type: this.$cookie.get("type"),
      docItems: [
        {
          title: "Chat",
          icon: "mdi-message-text",
          to: this.$cookie.get("previous-chat"),
        },
        { title: "Patients", icon: "mdi-clipboard-pulse", to: "/patients" },
        { title: "About", icon: "mdi-information", to: "/about" },
      ],
      patientItems: [
        {
          title: "Chat",
          icon: "mdi-message-text",
          to: this.$cookie.get("previous-chat"),
        },
        { title: "Doctors", icon: "mdi-doctor", to: "/doctors" },
        { title: "About", icon: "mdi-information", to: "/about" },
      ],
      adminItems: [
        {
          title: "Chat",
          icon: "mdi-message-text",
          to: this.$cookie.get("previous-chat"),
        },
        { title: "Admin", icon: "mdi-shield-account", to: "/admin" },
        { title: "Doctors", icon: "mdi-doctor", to: "/doctors" },
        { title: "Patients", icon: "mdi-clipboard-pulse", to: "/patients" },
        { title: "About", icon: "mdi-information", to: "/about" },
      ],
      notLogged: [
        {
          title: "Register",
          icon: "mdi-account-plus-outline",
          to: "/register",
        },
        { title: "About", icon: "mdi-information", to: "/about" },
      ],

      drawer: null,
    };
  },
  methods: {
    logout() {
      this.$cookie.delete("user");
      this.$cookie.delete("token");
      this.$cookie.delete("chatWith");
      this.$cookie.delete("previous-chat");
      this.$cookie.delete("type");
      this.$cookie.delete("avatar");
      this.$cookie.delete("patient");
      this.$cookie.delete("username");
      this.$cookie.delete("doctor");

      if (this.$route.name != "Welcome") {
        this.$router.push({ name: "Welcome" });
      }
      this.$router.go();
    },
  },
};
</script>