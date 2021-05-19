<template>
  <v-container>
    <div class="pa-6">
      <h1>CHAT</h1>
    </div>

    <ul id="list-of-messages" style="list-style-type: none">
      <li v-for="item in messages" :key="item.Message">
        [ {{ item.When }} ] <strong>{{ item.Name }}</strong
        >: {{ item.Message }}
      </li>
    </ul>

    <v-form>
      <v-container>
        <v-row>
          <v-col cols="8">
            <v-text-field
              v-model="message"
              :rules="msgRules"
              :counter="counter"
              required
              @keydown="sendEnter"
            ></v-text-field>
          </v-col>

          <v-col>
            <v-btn color="info" class="mr-4" @click="send">
              SEND
              <v-icon right dark> mdi-chat </v-icon>
            </v-btn>
          </v-col>
        </v-row>
      </v-container>
    </v-form>
  </v-container>
</template>

<script>
export default {
  data() {

    if (this.$cookie.get("user") == null) {
      return false
    }

    return {
      doctor: this.query,
      counter: 150,
      message: "",
      server: "192.168.0.14:8080",
      messages: [],
      msgRules: [
        (v) =>
          v.length <= this.counter ||
          "Message must be lesser than " + this.counter + " characters",
      ],
    };
  },
  props: ["query"],
  created() {
    console.log("creating the websocket ...");
    this.webSocket();
  },
  methods: {
    sendEnter: function (e) {
      // if pressed enter, just call the same method as from the send button
      if (e.keyCode === 13) {
        // Cancel the default action, if needed
        e.preventDefault();
        this.send();
      }
    },
    send: function () {
      if (!this.socket) {
        console.error("Error: There is no socket connection.");
        return false;
      }
      if (this.message !== "") {
        this.socket.send(JSON.stringify({ Message: this.message, Name: this.$cookie.get("user") }));
        this.message = "";
        return false;
      }
    },
    webSocket: function () {
      console.log(this.doctor);
      this.socket = new WebSocket("ws://" + this.server + "/ws/" + this.doctor);
      this.socket.onclose = function () {
        console.error("Connection has been closed.");
      };
      var self = this;
      this.socket.onmessage = function (e) {
        var msg = JSON.parse(e.data);
        self.messages.push(msg);
      };
    },
  },
};
</script>