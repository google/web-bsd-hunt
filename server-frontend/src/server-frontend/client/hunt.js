// Copyright 2016 The Web BSD Hunt Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
////////////////////////////////////////////////////////////////////////////////
//
// TODO: High-level file comment.
/*
 * debugging tools
 */
var SPINNER		= false;	// move a player (^ V < >) around the corners of the playfield
var PLAYER		= false;	// local playerInput handler instead of real one
var BEEPER		= false;	// periodically ring the bell
var DRAWDBG		= false;	// log playfield rendering commands
var INPUTDBG		= false;	// log game key events
var PAYLOADDBG		= false;	// log payloads sent to server
var GAMEDATADBG		= false;	// log gamedata payloads received from server
var DATAPARSEDBG	= false;	// log commands extracted from gamedata payload
var REPLYDBG		= false;	// log xmlhttprequest responseText
var RANDOM_PLAYFIELD	= false;	// set to true to fill the playfield with random characters for debugging
var COLOR		= "#FCFB78";	// text and border color.   This one chosen to match VT-220 yellowish / orange screen
//var COLOR		= "green";	// text and border color.   This one chosen as VT-220 green screen guess

function Player(screen, col, row) {
	this.pixi = new PIXI.Text("^", screen.textStyle);
	this.screen = screen;

	this.cmdPrefix = "move";
	this.cmdMap = {
		"K":	{text: "^"},
		"L":	{text: ">"},
		"J":	{text: "v"},
		"H":	{text: "<"},

		"k":	{x:0, y:-1},
		"l":	{x:1, y:0},
		"j":	{x:0, y:1},
		"h":	{x:-1, y:0}

	};

	this.col = col;
	this.row = row;

	this.pixi.position = screen.Position(this.col, this.row);

	this.keyCallback = function(cmd) {
		var op = this.cmdMap[cmd];
		if("text" in op) {
			this.pixi.text = op.text;
		}

		if("x" in op) {
			this.col += op.x;
			if(this.col < 0) {
				this.col = 0;
			}
			else
			if(this.col >= this.screen.dimensions.columns) {
				this.col = this.screen.dimensions.columns - 1;
			}
		}

		if("y" in op) {
			this.row += op.y;
			if(this.row < 0) {
				this.row = 0;
			}
			else
			if(this.row >= this.screen.dimensions.rows) {
				this.row = this.screen.dimensions.rows - 1;
			}
		}

		this.pixi.position = this.screen.Position(this.col, this.row);
	}.bind(this);

	this.FirstFrame = function() {
		return this.pixi;
	}.bind(this);

	this.NextFrame = function(now) {
		return this.pixi;
	}.bind(this);

	screen.stage.addChild(this.FirstFrame());
}

function Beeper(screen) {
	this.screen = screen

	this.lastUpdate		= 0;
	this.frequency		= 2500;	// ms between frames

	this.FirstFrame = function() {
		this.currentFrame = 0;
		this.lastUpdate = 0;
	}.bind(this);

	this.NextFrame = function(now) {
		var elapsed = (now - this.lastUpdate);

		if(elapsed < this.frequency) {
			return;
		}

		this.lastUpdate		= now;
		this.screen.BELL();
	}.bind(this);
}

function Spinner(screen) {
	this.lastUpdate		= 0;
	this.frequency		= 1000;	// ms between frames
	this.currentFrame	= 0;
	this.frames		= [ "^", ">", "v", "<" ];
	this.posDelta		= [
					/* top left */		screen.Position(0, 0),
					/* top right */		screen.Position(screen.dimensions.columns-1, 0),
					/* bottom right */	screen.Position(screen.dimensions.columns-1, screen.dimensions.rows-1),
					/* bottom left */	screen.Position(0, screen.dimensions.rows-1)
				  ],
	this.pixi		= new PIXI.Text(this.frames[this.currentFrame], screen.textStyle);
	this.pixi.position.x	= screen.cell.width;
	this.pixi.position.y	= screen.cell.height;
	this.pixi.scale.x	= 1;
	this.pixi.scale.y	= 1;

	this.FirstFrame = function() {
		this.currentFrame = 0;
		this.pixi.text = this.frames[this.currentFrame]

		return this.pixi
	}.bind(this);

	this.NextFrame = function(now) {
		var elapsed = (now - this.lastUpdate);

		if(elapsed < this.frequency) {
			return this.pixi;
		}

		this.lastUpdate		= now;
		this.currentFrame	= (this.currentFrame + 1) % this.frames.length;

		var pd = this.posDelta[this.currentFrame]
		this.pixi.text		= this.frames[this.currentFrame];
		this.pixi.position.x	= pd.x
		this.pixi.position.y	= pd.y

		return this.pixi;
	}.bind(this);

	screen.stage.addChild(this.FirstFrame());
}

function Help(screen) {
	var helpElement = document.getElementById("game-help");
	helpElement.style.marginLeft = screen.dimensions.width + 10;
	helpElement.style.height = screen.dimensions.height + 10;
	helpElement.style.width = 445;

	this.nlines = screen.dimensions.rows - 5;
	this.lastHeader = {};

	this.add = function(desc, str) {
		var h = {
			isHeader: false,
			desc: desc,
			str:  str
		};
		this.help.push(h);
	}.bind(this);

	this.addHeader = function(desc, str) {
		var h = {
			isHeader: true,
			desc: desc,
			str:  str
		};
		this.help.push(h);
	}.bind(this);

	this.help = [];
	this.addHeader("Object Identifiers",		"On Screen");
	this.add("Walls",				"- | +");
	this.add("Diagonal (deflecting) walls",		"/ \\");
	this.add("Doors (dispersion walls)",		"#");
	this.add("Small mine",				";");
	this.add("Large mine",				"g");
	this.add("Bullet",				":");
	this.add("Grenade",				"o");
	this.add("Satchel charge",			"O");
	this.add("Bomb",				"@");
	this.add("Small slime",				"s");
	this.add("Big slime",				"$");
	this.add("Me",					"> < ^ v");
	this.add("Other Players",			"} { i !");
	this.add("Explosion",				"*");
	this.add("Grenade & Large Mine Explosion",	"\\|/\n-*-\n/|\\");

	this.show = function() {
		var table = document.getElementById("game-help-table");
		while(table.firstChild) {
			table.removeChild(table.firstChild);
		}

		var tr  = document.createElement("tr");
		var th  = document.createElement("th");
		var h2  = document.createElement("h2");
		var help= document.createTextNode("Help (? for more)");
		var link= document.createElement("a");

		link.style.color = COLOR;
		link.target = "_blank";
		link.href = "http://www.unix.com/man-page/bsd/6/hunt/";
		link.appendChild(document.createTextNode("manpage"));

		th.colSpan = 2
		h2.appendChild(help);
		th.appendChild(link)
		th.appendChild(h2);
		tr.appendChild(th);

		table.appendChild(tr);

		//
		// handle wraparound first to deal with the case
		// where more help has been added since the last call.
		//
		if(this.helpStart >= this.help.length) {
			this.helpStart = 0;
		}

		var n = this.help.length - this.helpStart;
		if(n > this.nlines) {
			n = this.nlines;
		}
		for(i = 0; i < n; i++) {
			var tr  = document.createElement("tr");

			var h = this.help[this.helpStart + i];
			if(h.isHeader) {
				if(i > 0) {
					n = i;
					break;		// stop outputting at the next header
				}
				this.lastHeader = h;
			}

			if((i % this.nlines) == 0) {
				var th1 = document.createElement("th");
				var v1  = document.createTextNode(this.lastHeader.desc);
				th1.appendChild(v1);
				th1.innerHTML = th1.innerHTML.replace(/\n/g, "<br/>");
				tr.appendChild(th1);

				var th2 = document.createElement("th");
				var v2  = document.createTextNode(this.lastHeader.str);
				th2.appendChild(v2);
				th2.innerHTML = th2.innerHTML.replace(/\n/g, "<br/>");
				tr.appendChild(th2);
			}
			else {
				var td1 = document.createElement("td");
				var v1  = document.createTextNode(h.desc);
				td1.appendChild(v1);
				td1.innerHTML = td1.innerHTML.replace(/\n/g, "<br/>");
				tr.appendChild(td1);

				var td2 = document.createElement("td");
				var v2  = document.createTextNode(h.str);
				td2.appendChild(v2);
				td2.innerHTML = td2.innerHTML.replace(/\n/g, "<br/>");
				tr.appendChild(td2);
			}
			table.appendChild(tr);
		}

		this.helpStart += n;
	}.bind(this);

	this.FirstFrame = function() {
		this.helpStart = 0;
		this.show();
	}.bind(this);

	this.NextFrame = function() {
		this.show();
	}.bind(this);

	this.FirstFrame();
}

function Screen(font, bellSrc, rows, cols) {
	this.bell = new Audio(bellSrc); 
	this.muted = false;

	this.textStyle = {
		'font': font,
		'fill': COLOR,
		'align': 'left'
	};

	var t = new PIXI.Text("X", this.textStyle);

	this.cell = {
		width:	t.width,	// pixels
		height:	t.height,	// pixels
	};

	this.dimensions = {
		columns:cols,
		rows:	rows,
		width:	this.cell.width * (cols+1),
		height:	this.cell.height * (rows+1)
	};

	this.renderer	= new PIXI.autoDetectRenderer(this.dimensions.width, this.dimensions.height);
	this.stage	= new PIXI.Container();

	this.cursor = {
		column: 0,
		row: 0
	};

	this.randomContent = function() {
		var chars = [" ", "-", "|", "+", "/", "\\", "#", ";", "g", ":", "o", "O", "@", "s", "$", "}", "{", "i", "!", "*"];

		var i = 0
		if(Math.random() > 0.6) {
			i = Math.floor(Math.random() * chars.length);
		}

		return chars[i];
	}.bind(this);

	this.setup = function() {
		this.cells = [];
		for(row = 0; row < this.dimensions.rows; row++) {
			var line = [];
			for(col = 0; col < this.dimensions.columns; col++) {
				var content = " ";
				if(RANDOM_PLAYFIELD) {
					content = this.randomContent();
				}
				var cell = new PIXI.Text(content, this.textStyle);

				var pos = this.Position(col, row);
				cell.position.x	= pos.x;
				cell.position.y	= pos.y;

				line.push(cell);
				this.stage.addChild(cell);
			}
			this.cells.push(line);
		}
	}.bind(this);

	this.cursor.advance = function() {
		this.cursor.column++;
		if(this.cursor.column >= this.dimensions.columns) {
			this.cursor.column = 0;
			this.cursor.row++;
		}
		if(this.cursor.row >= this.dimensions.rows) {
			this.cursor.row = this.dimensions.rows - 1;	// can't advance past the last row
		}
	}.bind(this);

	this.cursor.move = function(row, column) {
		this.cursor.row = row;
		this.cursor.column = column;

		if(this.cursor.column >= this.dimensions.columns) {
			console.log("bad move column: " + column);
			this.cursor.column = this.dimensions.columns - 1;
		}
		if(this.cursor.row >= this.dimensions.rows) {
			console.log("bad move row: " + row);
			this.cursor.row = this.dimensions.rows - 1;
		}
	}.bind(this);

	this.Position = function(col, row) {
		var pos = {
			x: this.cell.width / 2,
			y: this.cell.height / 2,
		};
		if(col < 0 || col >= this.dimensions.columns) {
			console.log("Position: request col " + col + ": out of range 0.." + (this.dimensions.columns-1));
		}
		else {
			pos.x += col * this.cell.width;
		}

		if(row < 0 || row >= this.dimensions.rows) {
			console.log("Position: request row " + row + ": out of range 0.." + (this.dimensions.rows-1));
		}
		else {
			pos.y += row * this.cell.height;
		}

		if(pos.x >= this.dimensions.width || pos.y >= this.dimensions.height) {
			console.log("Position(" + row + "," + col + ") => offscreen " + pos);
		}

		return pos;
	}.bind(this);

	this.render = function() {
		this.renderer.render(this.stage);	// this is the main render call that makes pixi draw your container and its children.
	}.bind(this);

	//    Literal character	225 (ADDCH)
	//
	//	S: {uint8: 225} {uint8: c}
	//
	//	The client must draw the character with ASCII value c
	//	at the cursor position, then advance the cursor to the right.
	//	If the cursor goes past the rightmost column of the screen,
	//	it wraps, moving to the first column of the next line down.
	//	The cursor should never be advanced past the bottom row.
	//
	//	(ADDCH is provided as an escape prefix.)
	this.ADDCH = function(c) {
		if(DRAWDBG) {
			console.log("addch " + c);
		}
		this.cells[this.cursor.row][this.cursor.column].text = c;
		this.cursor.advance();
	}.bind(this);

	//    Cursor motion	237 (MOVE)
	//
	//	S: {uint8: 237} {uint8: y} {uint8: x}
	//
	//	The client must move its cursor to the absolute screen
	//	location y, x, where y=0 is the top of the screen and
	//	x=0 is the left of the screen.
	this.MOVE = function(row, column) {
		if(DRAWDBG) {
			console.log("move row " + row + " column " + column);
		}
		this.cursor.move(row, column);
	}.bind(this);

	//
	//    Refresh screen	242 (REFRESH)
	//
	//	S: {uint8: 242}
	//
	//	This indicates to the client that a burst of screen
	//	drawing has ended. Typically the client will flush its
	//	own drawing output so that the user can see the results.
	//
	//	Refreshing is the only time that the client must
	//	ensure that the user can see the current screen. (This
	//	is intended for use with curses' refresh() function.)
	this.REFRESH = function() {
		if(DRAWDBG) {
			console.log("refresh");
		}
		this.render();
	}.bind(this);

	//
	//    Clear to end of line 227 (CLRTOEOL)
	//
	//	S: {uint8: 227}
	//
	//	The client must replace all columns underneath and
	//	to the right of the cursor (on the one row) with 
	//	space characters. The cursor must not move.

	this.CLRTOEOL = function() {
		if(DRAWDBG) {
			console.log("clrtoeol");
		}
		for(var column = this.cursor.column; column < this.dimensions.columns; column++) {
			this.cells[this.cursor.row][column].text = " ";
		}
	}.bind(this);

	//
	//    End game		229 (ENDWIN)
	//
	//	S: {uint8: 229} {uint8: 32}
	//	S,C: <close>
	//
	//	S: {uint8: 229} {uint8: 236}
	//	S,C: <close>
	//
	//	The client and server must immediately close the connection.
	//	The client should also refresh the screen.
	//	If the second octet is 236 (LAST_PLAYER), then 
	//	the client should give the user an opportunity to quickly 
	//	re-enter the game. Otherwise the client should quit.
	this.ENDWIN = function(mode) {
		var msg = "Game Over!";
		switch(mode) {
		case 236:
			msg = "Game Over! (server exited)";
			if(DRAWDBG) {
				console.log("endwin: " + mode + ": client should quit");
			}
			break;
		case 32:
			msg = "You died! ctrl-j <enter> to re-join, ctrl-m {msg} <enter> to send message"
			if(DRAWDBG) {
				console.log("endwin: " + mode + ": client ask user to rejoin")
			}
			break;
		default:
			if(DRAWDBG) {
				console.log("endwin: " + mode + ": unhandled mode");
			}
			break;
		}

		this.MOVE(23, 0);
		this.addString(msg);
		this.CLRTOEOL();

/*
		this.MOVE(11, 33);
		this.addString("+------------+");
		this.MOVE(12, 33);
		this.addString("| Game Over! |");
		this.MOVE(13, 33);
		this.addString("+------------+");
*/
		this.REFRESH();
	}.bind(this);

	//
	//    Clear screen	195 (CLEAR)
	//
	//	S: {uint8: 195}
	//
	//	The client must erase all characters from the screen
	//	and move the cursor to the top left (x=0, y=0).
	this.CLEAR = function() {
		if(DRAWDBG) {
			console.log("clear");
		}
		for(row = 0; row < this.dimensions.rows; row++) {
			for(column = 0; column < this.dimensions.columns; column++) {
				this.cells[row][column].text = " ";
			}
		}
		this.MOVE(0, 0);
	}.bind(this);

	//
	//    Redraw screen	210 (REDRAW)
	//
	//	S: {uint8: 210}
	//
	//	The client should attempt to re-draw its screen.
	this.REDRAW = function() {
		if(DRAWDBG) {
			console.log("redraw");
		}
		this.REFRESH();
	}.bind(this);

	//
	//    Audible bell	226 (BELL)
	//
	//	S: {uint8: 226}
	//
	//	The client should generate a short audible tone for
	//	the user.
	this.BELL = function() {
		if(DRAWDBG) {
			console.log("bell");
		}
		if(!this.muted) {
			this.bell.play();
		}
	}.bind(this);

	//
	//    Server ready	231 (READY)
	//
	//	S: {uint8: 231} {uint8: n}
	//
	//	The client must refresh its screen.
	//
	//	The server indicates to the client that it has
	//	processed n of its characters in order, and is ready
	//	for more commands. This permits the client to 
	//	synchronise user actions with server responses if need be.
	this.READY = function(n) {
		if(DRAWDBG) {
			console.log("ready n " + n);
		}
		this.REFRESH();
	}.bind(this);

	//
	//    Characters other than the above.
	//
	//	S: {uint8: c}
	//
	//	The client must draw the character with ASCII value c
	//	in the same way as if it were preceded with ADDCH
	//	(see above).
	this.CHAR = function(c) {
		if(DRAWDBG) {
			console.log("char " + c);
		}
		this.ADDCH(c);
	}.bind(this);

	//
	// helper function to add strings to the screen
	//
	this.addString = function(str) {
		for(i = 0; i < str.length; i++) {
			this.ADDCH(str.charAt(i));
		}
	}.bind(this);

	this.drawEvents = {
		ADDCH:		225,	//    Literal character	225 (ADDCH)
		MOVE:		237,	//    Cursor motion	237 (MOVE)
		REFRESH:	242,	//    Refresh screen	242 (REFRESH)
		CLRTOEOL:	227,	//    Clear to end of line 227 (CLRTOEOL)
		ENDWIN:		229,	//    End game		229 (ENDWIN)
		CLEAR:		195,	//    Clear screen	195 (CLEAR)
		REDRAW:		210,	//    Redraw screen	210 (REDRAW)
		BELL:		226,	//    Audible bell	226 (BELL)
		READY:		231,	//    Server ready	231 (READY)
	};

	this.setup()
}

/*
 * this is a gross hack.
 */
function keyEventToString(keyEvent) {
	switch(keyEvent.code) {
	case "Digit4":
		if(keyEvent.shiftKey == true) {
			return "$";
		}
		return "4";
	case "Minus":
		if(keyEvent.shiftKey == true) {
			return "_";
		}
		return "-";
	case "Equal":
		if(keyEvent.shiftKey == true) {
			return "+";
		}
		return "=";
	case "Slash":
		if(keyEvent.shiftKey == true) {
			return "?";
		}
		return "/";
	}

	var key = String.fromCharCode(keyEvent.keyCode);
	if(!keyEvent.shiftKey) {
		key = key.toLowerCase();
	}

	return key;
}

/*
 * Is there a better way?
 */
function getRadioValue(name, defVal) {
	var buttons = document.getElementsByName(name);
	for(i = 0; i < buttons.length; i++) {
		if(buttons[i].checked) {
			return buttons[i].value;
		}
	}
	return defVal;
}

function setRadioValue(name, value) {
	var buttons = document.getElementsByName(name);

	for(var i = 0; i < buttons.length; i++) {
		if(buttons[i].value == value) {
			buttons[i].checked = true;
		} else {
			buttons[i].checked = false;
		}
	}
}

function Hunt(font, bellSrc, cols, rows) {
	this.screen = new Screen(font, bellSrc, rows, cols)

	this.playfield = document.getElementById("game-play");
	this.playfield.align = "center";
	this.playfield.style.verticalAlign = "middle";
	this.playfield.style.width = this.screen.dimensions.width + 10;
	this.playfield.style.height = this.screen.dimensions.height + 10;
	this.playfield.style.background = COLOR;
	this.screen.renderer.view.style.marginTop = "5px";
	this.playfield.appendChild(this.screen.renderer.view)

	this.helper = new Help(this.screen);

	if(SPINNER) {
		this.spinner = new Spinner(this.screen);
	}

	if(PLAYER) {
		this.player = new Player(this.screen, 0, 0);
	}

	if(BEEPER) {
		this.beeper = new Beeper(this.screen);
	}

	this.C_PLAYER	= 0	// response: game play tcp port
	this.C_MONITOR	= 1	// response: like C_PLAYER, but no response if 0 players
	this.C_MESSAGE	= 2	// response: number of players currently in the game
	this.C_SCORES	= 3	// response: statistics tcp port

	this.Q_CLOAK	= 1	// enter: cloaked
	this.Q_FLY	= 2	// enter: flying
	this.Q_SCAN	= 3	// enter: scanning

	this.stringToEnterStatus = function(str) {
		switch(str) {
		case "cloak":
			return this.Q_CLOAK;
		case "fly":
			return this.Q_FLY;
		case "scan":
			return this.Q_SCAN;
		}
		console.log("unknown enter status '" + str + "'. returning Q_FLY");
		return this.Q_FLY;
	}.bind(this);

	this.enterStatusToString = function(enter) {
		switch(enter) {
		case this.Q_CLOAK:
			return "cloak";
		case this.Q_FLY:
			return "fly";
		case this.Q_SCAN:
			return "scan";
		}

		console.log("unknown enter status '" + enter + "'. returning 'fly'");
		return "fly";
	}.bind(this);

	this.stringToTeam = function(str) {
		switch(str) {
		case "none":
		case "0":
		case "1":
		case "2":
		case "3":
		case "4":
		case "5":
		case "6":
		case "7":
		case "8":
		case "9":
			return str
		}
		console.log("bad team '" + str + "'. returning 'none'");
		return "none";
	}.bind(this);

	/*
	 * TODO(tad): handle quotes properly
	 */
	this.kvParse = function(s) {
		var results = [];
		if(s == "") {
			return results
		}

		var fields = s.split(",");
		for(var i = 0; i < fields.length; i++) {
			var key = fields[i];
			var val = "";

			var pos = key.indexOf("=");
			if(pos >= 0) {
				val = key.slice(pos+1);
				key = key.slice(0, pos);
			}

			var kv = {
				key: key,
				val: val
			};

			results.push(kv);
		}

		return results;
	}.bind(this);

	this.kvFind = function(s, key) {
		var k = this.kvParse(s);

		for(var i = 0; i < k.length; i++) {
			if(k[i].key == key) {
				return k[i];
			}
		}

		return null
	}.bind(this);

	this.kvToString = function(kv) {
		if(kv.val == "") {
			return kv.key
		}

		return kv.key + "=" + kv.val
	}.bind(this);

	this.kvalsToString = function(kvals) {
		var s = ""

		for(var i = 0; i < kvals.length; i++) {
			if(i > 0) {
				s = s + ","
			}
			s = s + this.kvToString(kvals[i])
		}

		return s
	}.bind(this);

	this.hashFind = function(key, defaultValue) {
		var hash = location.hash
		if(hash.indexOf("#") == 0) {
			hash = hash.slice(1);
		}

		var kv = this.kvFind(hash, key);
		if(kv == null) {
			return defaultValue
		}

		return kv.val
	}.bind(this);

	this.hashUpdate = function(key, val) {
		var hash = location.hash
		if(hash.indexOf("#") == 0) {
			hash = hash.slice(1);
		}

		var kvals = this.kvParse(hash);
		var i;
		for(i = 0; i < kvals.length; i++) {
			if(kvals[i].key == key) {
				kvals[i].val = val
				break
			}
		}
		if(i >= kvals.length) {
			var kv = {key: key, val: val};
			kvals.push(kv)
		}

		location.hash = this.kvalsToString(kvals);
	}.bind(this);

	this.setInstance = function(instance) {
		this.hashUpdate("instance", instance)
	}.bind(this);

	this.setName = function(name) {
		this.hashUpdate("name", name)
	}.bind(this);

	this.setTeam = function(team) {
		this.hashUpdate("team", team)
	}.bind(this);

	this.setEnterStatus = function(enter) {
		this.hashUpdate("enter", this.enterStatusToString(enter))
	}.bind(this);

	this.me = {
		Instance:	this.hashFind("instance", "0"),
		PlayerID:	"",
		Name:		this.hashFind("name", ""),
		Team:		this.stringToTeam(this.hashFind("team", "none")),
		EnterStatus:	this.stringToEnterStatus(this.hashFind("enter", "fly"))
	};

	this.log = function(text) {
		var elem = document.getElementById("game-log");
		elem.innerHTML = text;
	}.bind(this);

	this.animate = function() {
		requestAnimationFrame(this.animate);	// start the timer for the next animation loop

		var now = new Date().getTime();

		if(SPINNER) {
			this.spinner.NextFrame(now);
		}
		if(PLAYER) {
			this.player.NextFrame(now);
		}
		if(BEEPER) {
			this.beeper.NextFrame(now);
		}

		this.screen.render();	// this is the main render call that makes pixi draw your container and its children.
	}.bind(this);

	this.playerInputs = [
		{ key: "k", desc: "Move up" },
		{ key: "j", desc: "Move down" },
		{ key: "h", desc: "Move left" },
		{ key: "l", desc: "Move right" },

		{ key: "K", desc: "Face up" },
		{ key: "J", desc: "Face down" },
		{ key: "H", desc: "Face left" },
		{ key: "L", desc: "Face right" },

		{ key: "f", desc: "Fire a bullet (Takes 1 charge)" },
		{ key: "1", desc: "Fire a bullet (Takes 1 charge)" },			// alias
		{ key: "g", desc: "Throw grenade (Takes 9 charges)" },
		{ key: "2", desc: "Throw grenade (Takes 9 charges)" },			// alias
		{ key: "F", desc: "Throw satchel charge (Takes 25 charges)" },
		{ key: "3", desc: "Throw satchel charge (Takes 25 charges)" },		// alias
		{ key: "G", desc: "Throw bomb (Takes 49 charges)" },
		{ key: "4", desc: "Throw bomb (Takes 49 charges)" },			// alias
		{ key: "5", desc: "Throw big bomb (Takes 81 charges)" },
		{ key: "6", desc: "Throw even bigger bomb (Takes 121 charges)" },
		{ key: "7", desc: "Throw even more big bomb (Takes 169 charges)" },
		{ key: "8", desc: "Throw even more bigger bomb (Takes 225 charges)" },
		{ key: "9", desc: "Throw very big bomb (Takes 289 charges)" },
		{ key: "0", desc: "Throw very, very big bomb (Takes 361 charges)" },
		{ key: "@", desc: "Throw biggest bomb (Takes 441 charges)" },
		{ key: "o", desc: "Throw small slime (Takes 5 charges)" },
		{ key: "O", desc: "Throw big slime (Takes 10 charges)" },
		{ key: "p", desc: "Throw bigger slime (Takes 15 charges)" },
		{ key: "P", desc: "Throw biggest slime (Takes 20 charges)" },

		{ key: "s", desc: "Scan (show where other players are) (Takes 1 charge)" },
		{ key: "c", desc: "Cloak (hide from scanners) (Takes 1 charge)" },
	];

	this.gameInputs = [
		// game control
		{ key: "ctrl-l",	desc: "Redraw screen" },
		{ key: "ctrl-L",	desc: "Redraw screen" },			// alias
		{ key: "q",		desc: "Quit" },

		// custom additions
		{ key: "-",		desc: "Disable sound" },
		{ key: "+",		desc: "Enable sound" },
		{ key: "ctrl-j",	desc: "Join a game" },
		{ key: "ctrl-J",	desc: "Join a game" },				// alias
		{ key: "$",		desc: "Show Statistics" },
		{ key: "ctrl-m",	desc: "Message other players" },
		{ key: "ctrl-M",	desc: "Message other players" },		// alias
		{ key: "?",		desc: "Next page of help" }
	];

	this.processGameData = function(state) {
		if(GAMEDATADBG) {
			console.log(state);
		}
		if(state.Timeout) {
			console.log("timeout error: " + state.TimeoutError);
			return;
		}

		for(var i = 0; i < state.Data.length; ) {
			var c = state.Data[i];
			i++;

			var nremain = state.Data.length - i;

			if(DATAPARSEDBG) {
				console.log("Op[" + i + "] = " + c);
                        }
			switch(c) {
			case this.screen.drawEvents.ADDCH:
				if(nremain < 1) {
					console.log("addch need 1 have " + nremain);
					break
				}

				var txt = String.fromCharCode(state.Data[i]);
				i++;

				this.screen.ADDCH(txt)
				break;
			case this.screen.drawEvents.MOVE:
				if(nremain < 2) {
					console.log("move need 2 have " + nremain);
					break;
				}

				var y = state.Data[i];
				var x = state.Data[i+1];
				i += 2;

				this.screen.MOVE(y, x);
				break;
			case this.screen.drawEvents.REFRESH:
				this.screen.REFRESH();
				break;
			case this.screen.drawEvents.CLRTOEOL:
				this.screen.CLRTOEOL();
				break;
			case this.screen.drawEvents.ENDWIN:
				if(nremain < 1) {
					console.log("endwin need 1 have " + nremain);
					break;
				}
				var mode = state.Data[i];
				i++;
				this.screen.ENDWIN(mode);
				this.quitGame();
				break;
			case this.screen.drawEvents.CLEAR:
				this.screen.CLEAR();
				break;
			case this.screen.drawEvents.REDRAW:
				this.screen.REDRAW();
				break;
			case this.screen.drawEvents.BELL:
				this.screen.BELL();
				break;
			case this.screen.drawEvents.READY:
				if(nremain < 1) {
					console.log("ready need 1 have " + nremain);
					break;
				}
				var n = state.Data[i];
				i++;

				this.screen.READY(n);
				break;
			default:
				var txt = String.fromCharCode(c);
				this.screen.CHAR(txt)
			}
		}
	}.bind(this);

	this.sendPlayerKey = function(key) {
		if(this.me.PlayerID == "") {
			console.log("already quit")
			return
		}

		var payload = {
			PlayerID:	this.me.PlayerID,
			Keys:		key
		};

		var xhr = new XMLHttpRequest();

		xhr.open("PUT", "/api/v1/input/" + this.me.Instance, true);

		xhr.onload = function(e) {
			if(xhr.readyState != 4) {
				console.log("sendPlayerKey onload: readyState " + xhr.readyState);
				return
			}

			if(xhr.status != 200) {
				console.error("sendPlayerKey onload: " + xhr.status + ": " + xhr.statusText);
				return
			}
			if(INPUTDBG) {
				console.log("sendPlayerKey onload: ignore reply '" + xhr.responseText  +"'");
			}
		}.bind(this);

		xhr.onerror = function(e) {
			console.error("sendPlayerKey onerror: " + xhr.statusText);
		}.bind(this);

		xhr.ontimeout = function() {
			console.error("sendPlayerKey request timedout");
		}.bind(this);

		xhr.withCredentials = true;
		xhr.timeout = 60 * 1000;	/* ms */
		xhr.setRequestHeader("Content-Type", "application/json;charset=utf-8");
		xhr.setRequestHeader("Accept", "application/json;charset=utf-8");

		if(PAYLOADDBG) {
			console.log(payload)
		}

		xhr.send(JSON.stringify(payload));
	}.bind(this);

	//
	// side effect: clears this.me.PlayerID if the quit is successful
	//
	this.sendQuit = function() {
		if(this.me.PlayerID == "") {
			console.log("already quit")
			return
		}

		var payload = {
			PlayerID: this.me.PlayerID
		};

		var xhr = new XMLHttpRequest();

		xhr.open("PUT", "/api/v1/quit/" + this.me.Instance, true);

		xhr.onload = function(e) {
			if(xhr.readyState != 4) {
				console.log("onload: readyState " + xhr.readyState);
				return
			}

			if(xhr.status != 200) {
				console.error("onload: " + xhr.status + ": " + xhr.statusText);
			}

			if(REPLYDBG) {
				console.log("sendQuit onload: got '" + xhr.responseText  +"'");
			}

			var reply = JSON.parse(xhr.responseText)
			this.me.PlayerID = ""
		}.bind(this);

		xhr.onerror = function(e) {
			console.error("onerror: " + xhr.statusText);
		}.bind(this);

		xhr.ontimeout = function() {
			console.error("quit request timedout");
		}.bind(this);

		xhr.withCredentials = true;
		xhr.timeout = 5000;	/* ms */
		xhr.setRequestHeader("Content-Type", "application/json;charset=utf-8");
		xhr.setRequestHeader("Accept", "application/json;charset=utf-8");

		if(PAYLOADDBG) {
			console.log(payload)
		}

		xhr.send(JSON.stringify(payload));
	}.bind(this);

	this.sendGameData = function() {
		var payload = {
			PlayerID: this.me.PlayerID
		};

		var xhr = new XMLHttpRequest();
		xhr.open("PUT", "/api/v1/gamedata/" + this.me.Instance, true);

		xhr.onload = function(e) {
			if(xhr.readyState != 4) {
				console.log("onload: readyState " + xhr.readyState);
				return
			}

			switch(xhr.status) {
			default:
				console.error("onload: " + xhr.status + ": " + xhr.statusText);
				return
			case 408:
				/*
				 * this means there were no game state changes for a while, so just ask again
				 */
				if(REPLYDBG) {
					console.error("sendGameData onload: error " + xhr.status + ": " + xhr.statusText);
				}
				break
			case 200:
				/*
				 * got game state changes.  Process them then ask for more
				 */
				if(REPLYDBG) {
					console.log("sendGameData onload: got '" + xhr.responseText  +"'");
				}

				var data = JSON.parse(xhr.responseText)
				this.processGameData(data)
				break
			}

			this.sendGameData()
		}.bind(this);

		xhr.onerror = function(e) {
			console.error("onerror: [" + xhr.status + "]: " + xhr.statusText);
		}.bind(this);

		xhr.ontimeout = function() {
			console.error("timeout without game data");
			this.sendGameData()
		}.bind(this);

		xhr.withCredentials = true;
		xhr.timeout = 30*1000;	/* ms */
		xhr.setRequestHeader("Content-Type", "application/json;charset=utf-8");
		xhr.setRequestHeader("Accept", "application/json;charset=utf-8");

		if(PAYLOADDBG) {
			console.log(payload)
		}

		xhr.send(JSON.stringify(payload));
	}.bind(this);

	this.makejoin = function(name, team, enterStatus, connectMode) {
		var payload = {
			Uid:		777,
			Name:		name,
			Team:		team,
			EnterStatus:	enterStatus,
			Ttyname:	"web",
			ConnectMode:	connectMode,
		};

		return payload
	}

	this.sendJoin = function() {
		this.me.joining = true

		var payload = this.makejoin(this.me.Name, this.me.Team, this.me.EnterStatus, this.C_PLAYER)

		var xhr = new XMLHttpRequest();

		xhr.open("PUT", "/api/v1/join/" + this.me.Instance, true);

		xhr.onload = function(e) {
			if(xhr.readyState != 4) {
				console.log("onload: readyState " + xhr.readyState);
				return
			}

			this.me.joining = false

			if(xhr.status != 200) {
				console.error("onload: " + xhr.status + ": " + xhr.statusText);
				return
			}

			if(REPLYDBG) {
				console.log("sendJoin onload: got '" + xhr.responseText  +"'");
			}

			var reply = JSON.parse(xhr.responseText)
			this.me.PlayerID = reply.PlayerID

			this.sendGameData()
		}.bind(this);

		xhr.onerror = function(e) {
			this.me.joining = false
			console.error("onerror: " + xhr.statusText);
		}.bind(this);

		xhr.ontimeout = function() {
			this.me.joining = false
			console.error("join request timedout");
		}.bind(this);

		xhr.withCredentials = true;
		xhr.timeout = 5000;	/* ms */
		xhr.setRequestHeader("Content-Type", "application/json;charset=utf-8");
		xhr.setRequestHeader("Accept", "application/json;charset=utf-8");

		if(PAYLOADDBG) {
			console.log(payload)
		}

		xhr.send(JSON.stringify(payload));
	}.bind(this);

	this.sendMessage = function(name, team, message) {
		if(name == "") {
			name = "Anonymous";
		}

		var payload = {
			Join:		this.makejoin(name, team, this.Q_SCAN, this.C_MESSAGE),
			Message:	message
		};

		var xhr = new XMLHttpRequest();

		xhr.open("PUT", "/api/v1/message/" + this.me.Instance, true);

		xhr.onload = function(e) {
			if(xhr.readyState != 4) {
				console.log("onload: readyState " + xhr.readyState);
				return
			}
			if(xhr.status != 200) {
				console.error("onload: " + xhr.statusText);
				return
			}

			if(REPLYDBG) {
				console.log("sendMessage onload: got '" + xhr.responseText  +"'");
			}
		}.bind(this);

		xhr.onerror = function(e) {
			console.error("onerror: " + xhr.statusText);
		}.bind(this);

		xhr.ontimeout = function() {
			console.error("message request timedout");
		}.bind(this);

		xhr.withCredentials = true;
		xhr.timeout = 5000;	/* ms */
		xhr.setRequestHeader("Content-Type", "application/json;charset=utf-8")
		xhr.setRequestHeader("Accept", "application/json;charset=utf-8")

		if(PAYLOADDBG) {
			console.log(payload)
		}

		xhr.send(JSON.stringify(payload));
	}.bind(this);

	this.sendStats = function(instance) {
		var xhr = new XMLHttpRequest();

		xhr.open("PUT", "/api/v1/stats/" + instance, true);

		xhr.onload = function(e) {
			if(xhr.readyState != 4) {
				console.log("onload: readyState " + xhr.readyState);
				return
			}
			if(xhr.status != 200) {
				console.error("onload: " + xhr.status + ": " + xhr.statusText);
				return
			}

			if(REPLYDBG) {
				console.log("sendStats onload: got '" + xhr.responseText  +"'");
			}
			var msg = JSON.parse(xhr.responseText);
			this.log('<pre class="VT220" style="color:' + COLOR + ';">Stats for instance ' + instance + ':\n\n' + msg.Stats + '</pre>');
		}.bind(this);

		xhr.onerror = function(e) {
			console.error("onerror: " + xhr.statusText);
		}.bind(this);

		xhr.ontimeout = function() {
			console.error("stats request timedout");
		}.bind(this);

		xhr.withCredentials = true;
		xhr.timeout = 5000;	/* ms */
		xhr.setRequestHeader("Content-Type", "application/json;charset=utf-8")
		xhr.setRequestHeader("Accept", "application/json;charset=utf-8")

		xhr.send(null);
	}.bind(this);

	this.makeInstanceButton = function(instance) {
		var b = document.createElement("input");
		b.type = "button";
		b.value = instance;
		b.name = "instance-" + instance;
		b.onclick = function(event) {
			var input = document.getElementById("game-instance-id");
			if(!input.disabled) {
				input.value = event.target.value;
			}
			this.sendStats(event.target.value);
		}.bind(this);

		return b;
	}.bind(this);

	this.showInstanceButtons = function(instances) {
		var elem = document.getElementById("game-instance-buttons");
		elem.innerHTML = ""

		for(var i = 0; i < instances.length; i++) {
			elem.appendChild(instances[i]);
		}
	}.bind(this);

	this.sendGetInstances = function() {
		var xhr = new XMLHttpRequest();

		xhr.open("GET", "/api/v1/stats", true);

		xhr.onload = function(e) {
			if(xhr.readyState != 4) {
				console.log("onload: readyState " + xhr.readyState);
				return
			}
			if(xhr.status != 200) {
				console.error("onload: " + xhr.status + ": " + xhr.statusText);
				return
			}

			if(REPLYDBG) {
				console.log("sendGetInstances onload: got '" + xhr.responseText  +"'");
			}

			var msg = JSON.parse(xhr.responseText);

console.log(msg)

			var stats = "";

			this.input.instances = []
			if(msg.hasOwnProperty("AllStats") && Array.isArray(msg.AllStats)) {
				for(var i = 0; i < msg.AllStats.length; i++) {
					var instance = msg.AllStats[i]
					if(!instance.hasOwnProperty("InstanceID")) {
						continue;
					}

					var b = this.makeInstanceButton(instance.InstanceID);
					this.input.instances.push(b);

					var s = '<pre class="VT220" style="color:' + COLOR + ';">Stats for instance ' + instance.InstanceID + ':\n\n' + instance.Stats + '</pre>\n'
					stats = stats + s
				}
			}

			this.log(stats);
			this.showInstanceButtons(this.input.instances)
		}.bind(this);

		xhr.onerror = function(e) {
			console.error("onerror: " + xhr.statusText);
		}.bind(this);

		xhr.ontimeout = function() {
			console.error("stats request timedout");
		}.bind(this);

		xhr.withCredentials = true;
		xhr.timeout = 5000;	/* ms */
		xhr.setRequestHeader("Content-Type", "application/json;charset=utf-8")
		xhr.setRequestHeader("Accept", "application/json;charset=utf-8")

		xhr.send(null);
	}.bind(this);

	this.inputSetup = function(chat) {
		window.onkeydown = function(event) {
			/*
			 * Disable default delete behavior when in
			 * game mode (don't want an accidental keystroke to take
			 * the user away from the game), allow it to behave
			 * normaly when in message edit mode.
			 */
			if(!this.input.formMode && event.which == 8) { 
				event.preventDefault();   // turn off browser transition to the previous page 
			}
			return true;
		}.bind(this);

		var my_scope = this;

		this.input = {
			modifiers: {
				shifted:	false,
				controled:	false,
			},

			formMode:	false,
			keypresses: [
				{
					"this":		my_scope,
					"keys":		"shift",
					"on_keydown":	function() { this.input.modifiers.shifted = true; },
					"on_keyup":	function() { this.input.modifiers.shifted = false; },
				},
				{
					"this":		my_scope,
					"keys":		"control",
					"on_keydown":	function() { this.input.modifiers.controled = true; },
					"on_keyup":	function() { this.input.modifiers.controled = false; },
				},
			],

			callbacks: {}
		};
		this.input.modifiers.reset = function() {
			this.shifted = false;
			this.controled = false;
		}.bind(this.input.modifiers);

		/*
		 * trampoline reapplies modifiers and calls the user specified handler
		 */
		this.inputCallback = function(keyEvent, n, isRepeat) {
			key = keyEventToString(keyEvent);

			if(this.input.modifiers.shifted == true) {
				key = key.toUpperCase();
			}
			if(this.input.modifiers.controled == true) {
				key = "ctrl-" + key;
			}

			if(key in this.input.callbacks) {
				this.input.callbacks[key](key);
			}
		}.bind(this);

		this.registerKeyHandler = function(raw, callback) {
			key = raw.toLowerCase().replace("ctrl-", "");		// strip uppercase and control modifiers

			if(!(key in this.input.callbacks)) {			// tell the keypress library about each key only once
				var event = {
					"this":		my_scope,
					"keys":		key,
					"on_keydown":	this.inputCallback
				};
				this.input.keypresses.push(event);
			}
			this.input.callbacks[raw] = callback;
		}.bind(this);

		this.playerKeyCallback = function(cmd) {
			if(this.input.formMode) {
				console.log("ignore player key " + cmd);
				return;
			}
			if(INPUTDBG) {
				console.log("playerKey: " + cmd);
			}
			this.sendPlayerKey(cmd);
		}.bind(this);

		this.gameKeyCallback = function(cmd) {
			if(this.input.formMode) {
				console.log("ignore gamekey " + cmd);
				return;
			}

			if(INPUTDBG) {
				console.log("gameKey: " + cmd);
			}

			switch(cmd) {
			case "-":
				this.screen.muted = true;
				return;
			case "+":
				this.screen.muted = false;
				return;
			case "?":
				this.helper.NextFrame();
				return;
			case "ctrl-m":
			case "ctrl-M":
				this.input.chat.focus();
				return;
			case "ctrl-j":
			case "ctrl-J":
				this.input.login.focus();
				return;
			case "$":
				this.sendStats(this.me.Instance);
				return;
			case "q":
				this.quitGame();
				return;
			}
			console.log("gameKey '" + cmd + "' not handled")
		}.bind(this);

		this.loginCallback = function(e) {
			if(INPUTDBG) {
				console.log("login event: back to gameMode with name: " + this.input.login.value);
			}

			if(this.me.joining) {
				if(INPUTDBG) {
					console.log("join already in progress: ignore");
				}
			}
			else
			if(this.me.PlayerID != "") {
				if(INPUTDBG) {
					console.log("player already joined: ignore")
				}
			}
			else {
				if(INPUTDBG) {
					console.log("Player joining\n")
				}

				this.me.Instance = this.input.instance.value;
				this.me.Name = this.input.login.value;
				this.me.Team = this.stringToTeam(getRadioValue("game-login-team", this.me.Team));
				this.me.EnterStatus = this.stringToEnterStatus(getRadioValue("game-login-estatus", this.me.EnterStatus));

				this.setInstance(this.me.Instance);
				this.setName(this.me.Name);
				this.setTeam(this.me.Team);
				this.setEnterStatus(this.me.EnterStatus);

				if(this.me.Instance == "") {
					this.log("Instance required");

					e.preventDefault();
					return false;
				}

				if(this.me.Name == "") {
					this.log("Name required");

					e.preventDefault();
					return false;
				}

				this.log("")

				this.input.login.blur();
				this.input.instance.blur();

				this.sendJoin();

				this.input.login.disabled = true
				this.input.instance.disabled = true
			}

			e.preventDefault();
			return false;
		}.bind(this);

		this.quitGame = function() {
			this.sendQuit()
			this.input.login.disabled = false
			this.input.instance.disabled = false
			for(var i = 0; i < this.input.instances.length; i++) {
			}
		}.bind(this);

		this.messageCallback = function(e) {
			this.input.chat.blur();

			this.sendMessage(this.me.Name, this.me.Team, this.input.chat.value)
			this.input.chat.value = ""

			e.preventDefault();
			return false;
		}.bind(this);

		for(i = 0; i < this.playerInputs.length; i++) {
			if(PLAYER) {
				this.registerKeyHandler(this.playerInputs[i].key, this.player.keyCallback);
			}
			else {
				this.registerKeyHandler(this.playerInputs[i].key, this.playerKeyCallback);
			}
		}
		for(i = 0; i < this.gameInputs.length; i++) {
			this.registerKeyHandler(this.gameInputs[i].key, this.gameKeyCallback);
		}

		this.input.enableFormMode = function() {
			this.input.formMode = true;
			this.input.listener.stop_listening();
			this.input.modifiers.reset();
		}.bind(this);

		this.input.enableGameMode = function() {
			this.input.formMode = false;
			this.input.listener.listen();
		}.bind(this);

		this.input.listener = new window.keypress.Listener();
		this.input.listener.register_many(this.input.keypresses);
		this.input.listener.listen();

		this.input.chatform = document.getElementById("game-chat-form");
		this.input.chatform.addEventListener("submit", this.messageCallback);

		this.input.chat = document.getElementById("game-chat-message");
		this.input.chat.onfocus = function() {
			this.input.enableFormMode()
		}.bind(this);

		this.input.chat.onblur = function() {
			this.input.enableGameMode()
		}.bind(this);

		this.input.loginform = document.getElementById("game-login-form");
		this.input.loginform.addEventListener("submit", this.loginCallback);

		this.input.login = document.getElementById("game-login-name");
		this.input.login.value = this.me.Name
		this.input.login.onfocus = function() {
			this.input.enableFormMode()
		}.bind(this);

		this.input.login.onblur = function() {
			this.input.enableGameMode()
		}.bind(this);

		this.input.showInstances = document.getElementById("game-instance-show-all");
		this.input.showInstances.onclick = this.sendGetInstances;

		this.input.instanceform = document.getElementById("game-instance-form");
		this.input.instanceform.addEventListener("submit", this.loginCallback);

		this.input.instance = document.getElementById("game-instance-id");
		this.input.instance.value = this.me.Instance
		this.input.instance.onfocus = function() {
			this.input.enableFormMode()
		}.bind(this);

		this.input.instance.onblur = function() {
			this.input.enableGameMode()
		}.bind(this);

		setRadioValue("game-login-team", this.me.Team);
		setRadioValue("game-login-estatus", this.enterStatusToString(this.me.EnterStatus));

		this.sendGetInstances()
	}

	this.helper.addHeader("Player Commands", "Key");
	for(i = 0; i < this.playerInputs.length; i++) {
		var input = this.playerInputs[i];
		this.helper.add(input.desc, input.key);
	}

	this.helper.addHeader("Game Commands", "Key");
	for(i = 0; i < this.gameInputs.length; i++) {
		var input = this.gameInputs[i];
		this.helper.add(input.desc, input.key);
	}

	this.inputSetup();
	this.animate();
}

/*
 * Ugh:  Careful tweaking of font size and weight (including css) to make Chrome
 * on both Linux and Mac render the font correctly.  At "bold 20", it looks
 * great on Mac Chrome, but the tops of the letters are clipped on Linux Chrome.
 * at "bold 21", there are weird blank lines showing up inside the font.
 * "bold 20.5" seems to work fine.  At this size, on the Mac it looks slightly blurry
 * (like a real CRT!) and on Linux it's not clipped!
 */
var bellSrc = "/assets/sounds/Beep_Ping-SoundBible.com-217088958.wav";
var hunt = new Hunt('bold 20.5px Glass TTY VT220', bellSrc, 80, 24);
