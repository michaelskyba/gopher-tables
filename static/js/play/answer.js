let answer_request = new XMLHttpRequest()

document.onkeydown = function (e) {
	e = e || window.event

	// Pressed Enter
	if (e.keyCode == 13) {
		answer_request.open("POST", "/answer/", true)
		answer_request.send()
	}
}
