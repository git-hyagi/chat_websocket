const messages = new Vue({
    el: "#list-of-messages",
    data:
    {
        messages: [],
    }
})

const app = new Vue({
    el: "#userInput",
    data: {
        username: "John Doe",
        message: '',
        server: '192.168.0.14:8080',
    },
    methods: {
        send: function () {

            if (!this.socket) {
                console.log("Error: There is no socket connection.");
                return false;
            }

            if (this.message !== '') {
                this.socket.send(JSON.stringify({ "Message": this.message }));
                this.message = '';
                return false;
            }


        },
        webSocket: function () {
            this.socket = new WebSocket("ws://" + this.server + "/ws");
            this.socket.onclose = function () {
                console.log("Connection has been closed.")
            }
            this.socket.onmessage = function (e) {
                var msg = JSON.parse(e.data);
                messages.messages.push(msg);
            }
        }
    },

})

app.webSocket()