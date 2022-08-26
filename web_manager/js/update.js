$(document).ready(_ => {
	if (document.readyState != "complete") return

	let container = $("#list")

	let url = window.location.href
	url = new URL(url)
	c = url.searchParams.get("ws")
	c = c ? c : `ws://${window.location.host}/packet`

	console.log("connection to ws", c)

	let socket = new WebSocket(c)

	socket.onopen = _ => {
		socket.send("client: web;")
	}

	socket.onmessage = e => {
		let components = e.data.split(' ');
		const cmd = components.shift()
		const data = components.join(' ')

		switch (cmd) {
			case "packet":
				const json = JSON.parse(data)

				container.prepend(packet(json.Packet))
				break

			case "err":
				console.error("rec error: '%s'", e.data)
				return

			case "playerjoin":
				for (let name in data.split(",")) {
					addplayer(name)
				}
				return

			case "playerleave":
				for (let name in data.split(",")) {
					delplayer(name)
				}
				return

			case "pkts":
				data.split(",").forEach(pkt => !!pkt ? addpkt(pkt) : false)
				return

			case "hello":
				console.log("HELLO from server, version '%d'", data)
				return

			default:
				console.error("undefined action received, well idk now cmd: '%s'", cmd)
		}
	}

	socket.onclose = _ => {
		alert("ws closed, please reload")
	}

	socket.onerror = e => {
		console.error("websocket error", e.message)
	}
})

$(document).ready(function () {
	$('#playerselect').multiselect();
	$('#packetselect').multiselect();

	pktselector = $('#packetselect')
	playerselector = $('#playerselect')

	//addpkt("all")
	//addplayer(".all")

	playerselector.on("change", updatefilter)
	pktselector.on("change", updatefilter)
})


function updatefilter() {
	resetpkts()
	filterpkts(playerselector.val(), pktselector.val())
}

{ // pkts
	var pktselector
	var pkts = []

	function delpkt(name) {
		for (let i = 0; i < pkts.length; i++) {
			if (pkts[i] == name) {
				delete (pkts[i])
				updatepkt()
				return
			}

		}
	}

	function addpkt(name) {
		if (!pkts.includes(name)) {
			pkts.push(name)
			pkts.sort()

			pktselector.prepend(`<option value="${name}" selected>${name}</option>\n`)

			updatepkt()
			//PaymentRequestUpdateEvent
		}
	}

	function updatepkt() {
		pktselector.multiselect('rebuild');
	}
}

{ // players
	var playerselector
	var players = []

	function delplayer(name) {
		for (let i = 0; i < players.length; i++) {
			if (players[i] == name) {
				delete (players[i])
				updateplayers()
				return
			}

		}
	}

	function addplayer(name) {
		if (!players.includes(name)) {
			players.push(name)

			playerselector.append(`<option value="${name}" selected>${name}</option>\n`)

			updateplayers()
		}
	}

	function updateplayers() {
		playerselector.multiselect("rebuild")
	}
}

var packets = []

function packet(data) {
	packets.push(data)

	let content = []
	let icon, iconcolor

	switch (data.Type) {
		case "ToCltAORmAdd":
			content = [
				{ type: "txt", text: `rm: ${data.Cmd.Remove ? data.Cmd.Remove.length : 0}` },
				{ type: "txt", text: `add: ${data.Cmd.Add ? data.Cmd.Add.length : 0}` },
			]
			break

		case "ToCltHello":
			content = [
				{ type: "txt", text: `hello` },
			]
			break

		case "ToSrvNil":
			content = []
			break

		case "ToSrvInit":
			content = [
				{ type: "txt", text: `Serialisation ${data.Cmd.SerializeVer}`, title: "SerializeVer" },
				{ type: "txt", text: `<i class="bi bi-file-zip"></i>&nbsp;${data.Cmd.SupportedCompression}`, title: "SupportedCompression" },
				{ type: "txt", text: `proto: ${data.Cmd.MinProtoVer}-${data.Cmd.MaxProtoVer}`, title: "MinProtoVer, MaxProtoVer" },
			]
			break

		case "ToSrvFirstSRP":
			if (data.Cmd.EmptyPasswd) { // TODO: fixme
				icon = "key"
				iconcolor = "red"

				content = [
					{ type: "txt", text: `Empty Password` }
				]
			}
			break

		case "ToCltAcceptAuth":
			icon = "unlock"
			iconcolor = "green"

			content = [
				{ type: "txt", text: `<i class="bi bi-arrow-clockwise"></i>&nbsp;${data.Cmd.SendInterval}`, title: "SendInterval" },
				{ type: "txt", text: `<i class="bi bi-geo-alt"></i>&nbsp;${JSON.stringify(data.Cmd.PlayerPos)}`, title: "PlayerPos" },
				{ type: "txt", text: `<i class="bi bi-map"></i>&nbsp;${JSON.stringify(data.Cmd.MapSeed)}`, title: "MapSeed" },
			]
			break

		case "ToSrvInit2":
			content = [
				{ type: "txt", text: `<i class="bi bi-translate"></i>&nbsp;${escapeHTML(data.Cmd.Lang ? data.Cmd.Lang : "default")}`, title: "Lang" },
			]
			break

		case "ToCltItemDefs":
			content = [
				{ type: "txt", text: `#Defs: ${data.Cmd.Defs ? data.Cmd.Defs.length : 0}`, title: "Defs" },
				{ type: "txt", text: `#Aliases: ${data.Cmd.Aliases ? data.Cmd.Aliases.length : 0}`, title: "Aliases" },
			]
			break

		case "ToCltNodeDefs":
			content = [
				{ type: "txt", text: `#Defs: ${data.Cmd.Defs ? data.Cmd.Defs.length : 0}`, title: "Defs" },
			]
			break

		case "ToCltAnnounceMedia": // TODO split into categorys: sounds, mesh, texture, etc.
			content: [
				{ type: "txt", text: `#Files: ${data.Cmd.Files ? data.Cmd.Files.length : 0}`, title: "Files" },
			]
			break

		case "ToCltPrivs":
			content = [
				{ type: "txt", text: iconprivs(data.Cmd.Privs), title: "Privs" },
				{ type: "txt", text: escapeHTML(data.Cmd.Privs.join(", ")), title: "Privs" },
			]
			break

		case "ToCltDetachedInv": break; // TODO
		case "ToCltMovement": break; // TODO
		case "ToCltInvFormspec": break; // TODO

		case "ToCltHP":
			content = [
				{ type: "txt", text: `<i class="bi bi-heart-fill"></i>&nbsp;${data.Cmd.HP}</span>`, title: "HP" }
			]
			break

		case "ToCltBreath":
			content = [
				{ type: "txt", text: `<i class="bi bi-lungs"></i>&nbsp;${data.Cmd.Breath}</span>`, title: "Breath" }
			]
			break

		case "ToCltCSMRestrictionFlags": break; // TODO
		case "ToSrvReqMedia":
			icon = "file-earmark-plus"

			content = [
				{ type: "txt", text: `#Files: ${data.Cmd.Filenames ? data.Cmd.Files.Filenames : 0}`, title: "Filenames" }
			]
			break;

		case "ToCltMedia":
			icon = "file-earmark-arrow-down"

			content = [
				{ type: "txt", text: `#${data.Cmd.N}/${data.Cmd.I}`, title: "N/I" },
				{ type: "txt", text: `#Files: ${data.Cmd.Files ? data.Cmd.Files.length : 0}`, title: "Files" },
			]
			break;

		case "ToSrvCltReady":
			content = [
				{ type: "txt", text: `m.m.p.r: ${data.Cmd.Major},${data.Cmd.Minor}.${data.Cmd.Patch}.${data.Cmd.Reserved}`, title: "Major.Minor.Patch.Reserved" },
				{ type: "txt", text: `Version: ${escapeHTML(data.Cmd.Version)}`, title: "Version" },
				{ type: "txt", text: `Formspec: ${data.Cmd.Formspec}`, title: "Formspec" },
			]
			break;

		case "ToSrvPlayerPos":
			icon = "geo-alt"

			content = [
				{ type: "txt", text: `Pos: ${pos(data.Cmd.Pos.Pos100)}`, title: "Pos100" },
				{ type: "txt", text: `<i class="bi bi-speedometer"></i>&nbsp; ${pos(data.Cmd.Pos.Pos100)}`, title: "Vel100" },
				{ type: "txt", text: `Pitch100: ${data.Cmd.Pos.Pitch100}`, title: "Pitch100" },
				{ type: "txt", text: `Yaw100 ${data.Cmd.Pos.Yaw100}`, title: "Yaw100" },
				{ type: "txt", text: `FOV80 ${data.Cmd.Pos.FOV80}`, title: "FOV80" }, // TODO keys
			]
			break

		case "ToSrvSelectItem":
			content = [
				{ type: "txt", text: `Slot: ${data.Cmd.Slot}` },
			]

			break

		case "ToSrvInteract":
			icon = "hand-index-thumb"

			content = [
				{ type: "txt", text: `${data.Cmd.Action}` },
				{ type: "txt", text: `Slot: ${data.Cmd.ItemSlot}` },
				{ type: "txt", text: `Keys: ${keysString(data.Cmd.Pos.Keys).join(", ")}` }, // TODO
			]

			break

		case "ToSrvFallDmg":
			icon = "hammer"
			iconcolor = "red"

			content = [
				{ type: "txt", text: `<i class="bi bi-heart-fill"></i>&nbsp;${data.Cmd.Amount}`, title: "Amount" },
			]
			break;

		case "ToSrvInvAction":
			icon = "archive"

			content = [
				{type:"txt", text: data.Cmd.Action, title: "Action"}
			]
			break

		case "ToCltKick":
			icon = "box-arrow-left"

			content = [
				{type:"txt", text: data.Cmd.Reason, title:"Reason"},
//				{type:"txt", text: data.Cmd.Custom, title:"Custom"},
//				{type:"txt", text: data.Cmd.Reconnect, title:"Reconnect"},
			]

			if(data.Cmd.Custom)
				content.push({type:"txt", text: data.Cmd.Custom, title:"Custom"})

			if(data.Cmd.Reconnect)
				content.push({type:"icon", data: "arrow-clockwise"})

			break

		case "ToCltBlkData":
			icon = "box"

			content = [ // TODO: server side send pos
				{type:"txt", text: `<i class="bi bi-geo-alt">`},
			]
			break

		case "ToSrvGotBlks":
			// ok
			break

		default:
			return data.pntr = genData({
				type: "err",
				name: data.Name,
				srv: data.Srv,
				clt: data.Clt,
				content: [
					{ type: "txt", text: "type not parsed" },
					{ type: "txt", text: data.Type },
					{ type: "txt", text: JSON.stringify(data).substring(0, 500) },
				],
			})
	}

	console.log(content)
	updatefilter()

	return data.pntr = genData({
		type: "pkt",
		name: data.Name,
		srv: data.Srv,
		clt: data.Clt,
		icon: icon,
		iconcolor: iconcolor,
		content: [
			{ type: "txt", text: data.Type },
			...content,
		],
	})
}

function filterpkts(names = [], types = []) {
	packets.forEach(pkt => {
		if(!pkt.pntr) return

		if (
			(types.includes("all") || types.includes(pkt.Type))
			&& (names.includes(".all") || names.includes(pkt.Name))
		) {
			pkt.pntr.style = "visibility:visible; height:auto; margin-bottom:1rem !important"
		} else {
			pkt.pntr.style = "visibility:hidden; height: 0px; margin-bottom:0 !important"
		}
	})
}

function resetpkts() {
	packets.forEach(pkt => pkt.pntr ? pkt.pntr.style = "visibility:visible" : undefined)
}

{ // parsing helpers
	keys = {
		ForwardKey: 1 << 1,
		BackwardKey: 1 << 2,
		LeftKey: 1 << 3,
		RightKey: 1 << 4,
		JumpKey: 1 << 5,
		SpecialKey: 1 << 6,
		SneakKey: 1 << 7,
		DigKey: 1 << 8,
		PlaceKey: 1 << 9,
		ZoomKey: 1 << 10,
	}

	// https://pkg.go.dev/github.com/anon55555/mt#Keys
	function keysString(k) {
		arr = []

		for (key in keys) {
			if (k & keys[key]) {
				arr.push(key)
			}
		}

		return arr
	}

	// thanks to: Vitim.us for https://stackoverflow.com/a/22706073
	function escapeHTML(str) {
		var p = document.createElement("p");
		p.appendChild(document.createTextNode(str));
		return p.innerHTML;
	}

	function pos(arg) {
		if (arg == undefined) {
			console.error("arg is undefined, and thats bad")
			return "err"
		}

		return [arg[0] / 100, arg[1] / 100, arg[2] / 100]
	}

	let privicos = { "fly": "airplane", "interact": "hand-index-thumb", "fast": "fast-forward" }

	function iconprivs(privs) {
		let str = ""

		privs.filter(p => {
			let ico = privicos[p]
			if (ico) {
				str += `<i class="bi bi-${ico}"></i>&nbsp;`
				return false
			} else {
				return true
			}
		})

		return str
	}

	function genData(data) {
		let obj = document.createElement("div")
		obj.classList = "input-group mb-3"

		try {
			switch (data.type) {
				case undefined:
					return "err"

				case "pkt":
					obj.innerHTML = `${data.icon ? formatContent({ type: "icon", data: data.icon, color: data.iconcolor }) : '<span class="input-group-text" id="basic-addon1"><i class="bi bi-box-seam"></i></span>'}
    ${username(data)}
    ${data.content.map(entry => formatContent(entry)).join("")}`
					break

				case "msg":
					obj.innerHTML = `<span class="input-group-text" id="basic-addon1"><i class="bi bi-chat-left-dots"></i></span>
${data.content.map(entry => formatContent(entry)).join("")}`
					break

				case "raw":
					obj.innerHTML = data.content.map(entry => formatContent(entry)).join("")
					break

				case "err":
					obj.innerHTML = `<span class="input-group-text" id="basic-addon1"><i class="bi bi-exclamation-circle" style="color:red"></i></span>
${username(data)}
${data.content.map(entry => formatContent(entry)).join("")}`
			}
		} catch (e) {
			obj.innerHTML = "err " + e
			return obj
		}

		return obj
	}

	function username(data) {
		if (data.clt == "") return

		if (data.srv == true) {
			return `<span class="input-group-text" id="basic-addon1" title="${data.clt}"><i class="bi bi-hdd"></i>&nbsp;-&gt;&nbsp;<i class="bi bi-pc-display"></i>&nbsp;${escapeHTML(data.name)}</span>`
		} else {
			return `<span class="input-group-text" id="basic-addon1"><i title="${data.clt}" class="bi bi-pc-display">&nbsp;-&gt;&nbsp;<i class="bi bi-hdd"></i></i>&nbsp;${escapeHTML(data.name)}</span>`
		}
	}

	function formatContent(data) {
		switch (data.type) {
			case "txt":
				return `<span class="input-group-text" id="basic-addon1" style="color:${data.color}" ${data.title ? "title=\"" + data.title + "\"" : ""}>${data.text}</span>`

			case "icon":
				return `<span class="input-group-text" id="basic-addon1" style="color:${data.color}"><i class="bi bi-${data.data}"></i></span>`
		}

	}
}