let answer_request = new XMLHttpRequest()
let answer_input = document.getElementById("answer")

// Initiate the game once both players have joined
function initiate() {
	let init_request = new XMLHttpRequest()
	init_request.onload = function() {
		document.getElementById("question").innerHTML = this.responseText
		document.getElementById("submit").style.display = "block"
	}

	init_request.open("GET", `/init_question/`, true)
	init_request.send()
}

document.onkeydown = function (e) {
	e = e || window.event

	// Pressed Enter
	if (e.keyCode == 13) {
		let answer = answer_input.value
		answer_input.value = ""

		answer_request.open("GET", `/answer/${answer}/`, true)
		answer_request.send()
	}
}
