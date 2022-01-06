let answer_request = new XMLHttpRequest()
let answer_input = document.getElementById("answer")

document.onkeydown = function (e) {
	e = e || window.event

	// Pressed Enter
	if (e.keyCode == 13) {
		let answer = answer_input.value
		answer_input.value = ""

		answer_request.open("POST", `/answer/${game_id}/${answer}/`, true)
		answer_request.send()
	}
}
