// "Refresh"
document.getElementById("refresh").onclick = function() {
	location.reload()
}

// "Create game"
document.getElementById("create").onclick = function() {
	window.location.replace("/create/")
}

// Individual game buttons
let names   = document.getElementsByClassName("join_name")
let buttons = document.getElementsByClassName("join_button")
for (let i = 0; i < buttons.length; i++) {
	buttons[i].onclick = function() {
		window.location.replace(`/join/${names[i].innerHTML}/`)
	}
}
