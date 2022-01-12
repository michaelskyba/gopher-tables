let answer_request = new XMLHttpRequest()
answer_request.onerror = function(){location.reload()}

let answer_input = document.getElementById("answer")
let wrong_streak = 0
let original_placeholder = document.getElementById("answer").placeholder

// Initiate the game once both players have joined
function initiate() {
	let init_request = new XMLHttpRequest()
	int_request.onerror = function(){location.reload()}
	init_request.onload = function() {
		document.getElementById("question").innerHTML = this.responseText
		document.getElementById("submit").style.display = "block"
	}

	init_request.open("GET", `/init_question/`, true)
	init_request.send()
}

// After submitting an answer and receiving a response from /answer/
answer_request.onload = function() {

	if (this.responseText.trim() == "incorrect") {
		wrong_streak++

		let placeholder = "Incorrect answer"
		if (wrong_streak > 1) {
			placeholder = `${placeholder} (x${wrong_streak})`
		}
		document.getElementById("answer").placeholder = placeholder
	}

	else {
		wrong_streak = 0
		document.getElementById("answer").placeholder = original_placeholder
		document.getElementById("question").innerHTML = this.responseText
	}
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
