<template>
<v-container>
<div class="pa-6">
    <h1>DOCTORS</h1>
    </div>
  <v-card class="mx-auto" max-width="500">
    <!--
    <v-toolbar
      color="deep-purple accent-4"
      dark
    >
      <v-app-bar-nav-icon></v-app-bar-nav-icon>

      <v-toolbar-title>New Chat</v-toolbar-title>

      <v-spacer></v-spacer>

      <v-btn icon>
        <v-icon>mdi-magnify</v-icon>
      </v-btn>
    </v-toolbar>
    -->

    <v-list subheader>
      <v-subheader>Recent chat</v-subheader>

      <v-list-item v-for="chat in recent" :key="chat.title" :to="chat.to">
        <v-list-item-avatar>
          <v-img :alt="`${chat.title} avatar`" :src="chat.avatar"></v-img>
        </v-list-item-avatar>

        <v-list-item-content>
          <v-list-item-title v-text="chat.title"></v-list-item-title>
        </v-list-item-content>

        <v-list-item-icon>
          <v-icon :color="chat.active ? 'deep-purple accent-4' : 'grey'">
            mdi-message-outline
          </v-icon>
        </v-list-item-icon>
      </v-list-item>
    </v-list>

    <v-divider></v-divider>

    <v-list subheader>
      <v-subheader>Previous chats</v-subheader>

      <v-list-item v-for="chat in previous" :key="chat.title" :to="chat.to">
        <v-list-item-avatar>
          <v-img :alt="`${chat.title} avatar`" :src="chat.avatar"></v-img>
        </v-list-item-avatar>

        <v-list-item-content>
          <v-list-item-title v-text="chat.title"></v-list-item-title>
        </v-list-item-content>
      </v-list-item>
    </v-list>
  </v-card>
  </v-container>
</template>


<script>
export default {
  data: () => ({
    recent: [
      {
        active: true,
        avatar: "https://cdn.vuetifyjs.com/images/lists/1.jpg",
        title: "Muleque Xandy",
        to: "/chat?q=xandy",
      },
      {
        active: true,
        avatar: "https://cdn.vuetifyjs.com/images/lists/2.jpg",
        title: "Patch Adams",
        to: "/chat?q=padams",
      },
      {
        avatar: "https://cdn.vuetifyjs.com/images/lists/4.jpg",
        title: "Nazi Yamauchi (cloroquina)",
        to: "/chat?q=charlata",
      },
      {
        avatar: "https://cdn.vuetifyjs.com/images/lists/3.jpg",
        title: "Elizabeth Blackwell",
        to: "/about",
      },
    ],
    previous: [
      {
        title: "Marya Árvore de Natal",
        avatar: "https://cdn.vuetifyjs.com/images/lists/5.jpg",
        to: "/chat?q=cloroquina-captain",
        default: true,
      },
    ],
  }),
};
</script>