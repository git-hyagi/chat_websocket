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
            ></v-text-field>
          </v-col>

          <v-col>
            <v-btn class="mr-4" @click="send">SEND</v-btn>
          </v-col>
        </v-row>
      </v-container>
    </v-form>
  </v-container>
</template>

<script>
export default {
  data() {
    return {
      doctor: "",
      username: "John Doe",
      counter: 150,
      message: "",
      server: "192.168.0.14:8080",
      messages: [],
      msgRules: [
        (v) =>
          v.length <= this.counter ||
          "Message must be lesser than " + this.counter + " characters",
      ],
      server: "192.168.0.14:8080",
    };
  },
  created() {
    console.log("creating the websocket ...");
    this.webSocket();
  },
  methods: {
    send: function () {
      if (!this.socket) {
        console.error("Error: There is no socket connection.");
        return false;
      }
      if (this.message !== "") {
        this.socket.send(JSON.stringify({ Message: this.message }));
        this.message = "";
        return false;
      }
    },
    webSocket: function () {
      this.socket = new WebSocket("ws://" + this.server + "/ws");
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