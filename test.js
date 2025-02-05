const socket = new WebSocket('ws://localhost:8080/ws');

let messagesSent = 0;
const rooms = ['general', 'sub_general'];

function generateRoomName() {
    return 'This is random message to room general' + Math.random().toString(36).substring(2, 15);
}

function pickRandomRoom() {
  return rooms[Math.floor(Math.random() * rooms.length)];
}

function joinRooms() {
    rooms.forEach(room => {
        const createMessage = {
            command: 'join_room',
            room: room
        };
        socket.send(JSON.stringify(createMessage));
    });
}

function sendMessage(mes) {
  const message = {
    command: 'send_message',
    room: pickRandomRoom(),
    message: mes
  }

  socket.send(JSON.stringify(message));
}

socket.addEventListener('open', (event) => {
    console.log('Connected to WebSocket server');

    joinRooms();
    const interval = setInterval(() => {
        if (messagesSent < 10000) {
            sendMessage(generateRoomName());
            messagesSent++;
        } else {
            clearInterval(interval);
            console.log('Finished creating 200 rooms');
            socket.close();
        }
    }, 1);     

    
});

socket.addEventListener('message', (event) => {
    const response = JSON.parse(event.data);
    console.log(`Messages sent ${messagesSent}: ${response.message} to room: ${response.room}`);
});

socket.addEventListener('error', (event) => {
    console.error('WebSocket error:', event);
});

socket.addEventListener('close', (event) => {
    console.log('Disconnected from WebSocket server');
});
