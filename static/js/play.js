let xhr = new XMLHttpRequest()
xhr.open("GET", "/progress/", true)

xhr.onload = function() {
	console.log(this.status)
	console.log(this.responseText)
}

xhr.send()
