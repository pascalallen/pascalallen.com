$(function() {

    var projectAPIKey = "DAK6f263c3f51a44cde9d8c5ee500877dcf";
    var username = "user1@smssend.yahoo.com";
    var password = "1quaeratestdolore1";
  
    kandy.setup({

        listeners: {
            message: onMessageReceived
        }
    });

    kandy.login(projectAPIKey, username, password, onLoginSuccess, onLoginFailure);

    function onLoginSuccess() {
        console.log("Login was successful.");
    }

    // What to do on a failed login.
    function onLoginFailure() {
        console.log("Login failed. Make sure you input the user's credentials!");
    }

    /**
     * Called when the `message` event is triggered.
     * Receives the message object as a parameter.
     */
    function onMessageReceived(message) {
        // Display the message as incoming.
        var sender = message.sender.user_id;
        // Create the message element. Use Lodash to escape the message for security purposes.
        var element = "<div>Incoming (" + sender + "): " + _.escape(message.message.text) + "</div>"
        document.getElementById("chat-messages").innerHTML += element;
    }
  
    // event handler for send message button
    $('#send-btn').on('click', function() {
        var sender = $('#sms_from').val();
        var receiver = $('#sms_to').val();
        var message = $('#sms_message').val();

        kandy.messaging.sendSMS(
            receiver,
            sender,
            message,
            function() {
                alert('sms sent');
            },
            function(message, status) {
                alert(message + status);
        });
    });
  
  });
