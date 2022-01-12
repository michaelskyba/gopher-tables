// Question initialization
let started = true
if (initial_question == "") {
	started = false
	document.getElementById("question").innerHTML = "Waiting for a player..."
	document.getElementById("submit").style.display = "none"
}
else {
	document.getElementById("question").innerHTML = initial_question
}

const progress_url = `/progress/${game_id}/`

let progress_request = new XMLHttpRequest()
progress_request.onload = function() {

	// TODO: Check for 200 (OK) status using this.status

	let scores = JSON.parse(this.responseText)

	// I'm assuming the order will be the same every time
	// If not, I'll have to sort alphabetically to keep the locations consistent
	Object.keys(scores).forEach((name, i) => {

		let score = scores[name]

		document.getElementById(`name_${i}`).innerHTML = name
		document.getElementById(`progress_${i}`).value = score

		// We can start, since everyone has joined
		if (i > 0 && started == false) {
			started = true
			initiate()
		}

		// Somebody won, so the game is over
		if (score > 9) {
			document.getElementById("question").innerHTML = `${name} is the winner!`
			document.getElementById("submit").style.display = "none"
			clearInterval(interval)
		}
	})
}

// Constantly check for updates in score
let interval = setInterval(function() {
	progress_request.open("GET", progress_url, true)
	progress_request.send()

}, 100)
