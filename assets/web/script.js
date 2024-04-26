let displayCount = 0;
const displayTimeout = 60 * 1000;

function showMessage(chatMessage) {
	let container = document.getElementById("chat-overlay");
	let div = document.createElement("div");
	// Format message
	div.innerHTML = `
		<span class="author">${chatMessage.author}:</span>
		<span class="text">${chatMessage.text}</span>
	`;
	div.setAttribute("class", `message ${chatMessage.platform.toLowerCase()}`);
	container.appendChild(div);

	setTimeout(function() { container.removeChild(div); }, displayTimeout);
	window.scrollTo(0, document.body.scrollHeight);
}

async function loadMessages() {
	const resp = await fetch("/chat");
	let chatHistory = await resp.json();
	for (let i=displayCount+1; i<chatHistory.length; i++) {
		showMessage(chatHistory[i]);
		displayCount=i;
	}
}

showMessage({
	"author": "LiveQuest",
	"text": "Pronto para receber mensagens",
	"platform": ""
});

setInterval(loadMessages, 500);
