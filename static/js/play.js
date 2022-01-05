let scores

let xhr = new XMLHttpRequest()
xhr.open("GET", "/progress/", true)

xhr.onload = function() {
	// console.log(this.status)

	scores = JSON.parse(this.responseText)
	console.log(scores)
}

let interval = setInterval(function() {
	xhr.open("GET", "/progress/", true)
	xhr.send()

	console.log(scores)

}, 100)
