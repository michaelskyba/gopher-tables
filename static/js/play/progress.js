const progress_url = `/progress/${game_id}/`

let progress_request = new XMLHttpRequest()
progress_request.onload = function() {

	// TODO: Check for 200 (OK) status using this.status

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
	progress_request.open("GET", progress_url, true)
	progress_request.send()

}, 100)
