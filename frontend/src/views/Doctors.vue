<template>
  <v-container>
    <h1 class="pa-6">Doctors</h1>
    <v-card max-width="500" class="mx-auto">
      <v-list>
        <!-- <v-list-item v-for="item in items" :key="item.title"  :to="item.to"> -->
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
  data() {
    if (this.$cookie.get("user") == null) {
      return false;
    }

    return {
      items: [
        {
          title: "Muleque Xandy",
          subtitle: "Ortopedia",
          avatar: "https://cdn.vuetifyjs.com/images/lists/1.jpg",
          to: "/chat?q=xandy",
        },
        {
          title: "Patch Adams",
          subtitle: "Clínica",
          avatar: "https://cdn.vuetifyjs.com/images/lists/2.jpg",
          to: "/chat?q=padams",
        },
        {
          title: "Nazi Yamauchi",
          subtitle: "Charlatã Minion",
          avatar: "https://cdn.vuetifyjs.com/images/lists/3.jpg",
          to: "/chat?q=charlata",
        },
        {
          title: "Marya Árvore de Natal",
          subtitle: "Minion",
          avatar: "https://cdn.vuetifyjs.com/images/lists/5.jpg",
          to: "/chat?q=cloroquina-captain",
        },
        {
          title: "Elizabeth Blackwell",
          subtitle: "",
          avatar: "https://cdn.vuetifyjs.com/images/lists/4.jpg",
          to: "/about",
        },
      ],
    };
  },
  methods: {
    updatePrevious: function (item) {
      this.$cookie.set("previous-chat", item);
      this.$router.push(item);
    },
  },
};
</script>