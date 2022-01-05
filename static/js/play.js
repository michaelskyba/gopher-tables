const progress_url = `/progress/${game_id}/`

let xhr = new XMLHttpRequest()

xhr.onload = function() {
	// console.log(this.status)

	let scores = JSON.parse(this.responseText)

	// I'm assuming the order will be the same every time
	// If not, I'll have to sort alphabetically to keep the locations consistent
	Object.keys(scores).forEach((name, i) => {
		document.getElementById(`name_${i}`).innerHTML = name
		document.getElementById(`progress_${i}`).value = scores[name]
	})
}

// Constantly check for updates in score
let interval = setInterval(function() {
	xhr.open("GET", progress_url, true)
	xhr.send()

}, 100)
