let displayCount = 0;

function showMessage(chatMessage) {
	let container = document.getElementById("chat-overlay");
	let div = document.createElement("div");

	// Format message
	div.innerHTML = `
		<span class="author">${chatMessage.author}:</span>
		<span class="text">${chatMessage.text}</span>
	`;
	div.setAttribute("class", "message");
	container.appendChild(div);

	setTimeout(function() {
		container.removeChild(div);
	}, 20 * 1000);
	window.scrollTo(0, document.body.scrollHeight);
}

async function loadMessages() {
	const resp = await fetch("/chat");
	let gameState = await resp.json();
	console.log(gameState);
	for (var i=displayCount+1; i<gameState.chatHistory.length; i++) {
		showMessage(gameState.chatHistory[i]);
		displayCount=i;
	}
}

showMessage({
	"author": "LiveQuest",
	"text": "Pronto para receber mensagens"
});

setInterval(loadMessages, 500);
