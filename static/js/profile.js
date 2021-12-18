let level = 1
let threshold = 10

while (wins >= threshold) {
	wins -= threshold
	threshold = Math.ceil(1.1 * threshold)
	level++
}
let progress = Math.floor(wins / threshold * 100)

document.getElementById("percent").innerHTML = progress
document.getElementById("progress").value = progress
document.getElementById("level").innerHTML = level
