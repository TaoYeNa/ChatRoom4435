
const LEFT = "left";
const RIGHT = "right";

const EVENT_MESSAGE = "message"
const EVENT_OTHER = "other"

const userPhotos = [
    "../../img/p1.svg",
    "../../img/p2.svg",
    "../../img/p3.svg",
    "../../img/p4.svg",
    "../../img/p5.svg",
    "../../img/p6.svg",
    "../../img/p7.svg",
]
var PERSON_IMG = userPhotos[getRandomNum(0, userPhotos.length - 1)];
var PERSON_NAME = "Guest" + Math.floor(Math.random() * 1000);

var port;
var ws;


var name = "Guest" + Math.floor(Math.random() * 1000);
var chatroom = document.getElementsByClassName("msger-chat")
var text = document.getElementById("msg");
var image_up = document.getElementById("image");
var send = document.getElementById("send")

function buttonHandle(obj){
    port = obj.id
    var url = "ws://localhost:" + port + "/ws?id=" + PERSON_NAME;
    ws = new WebSocket(url);
    
	$(".cover").hide();
	$(".port-selection").hide();
	ws.onmessage = function (e) {
	    var m = JSON.parse(e.data)
	    var msg = ""
	    switch (m.event) {
		case EVENT_MESSAGE:
		    if (m.name == PERSON_NAME) {
		        if(m.img_64 != null){
		            msg = getMessage(m.name, m.photo, RIGHT, m.content ,m.img_64);
		        }
		        else{
		            msg = getMessage(m.name, m.photo, RIGHT, m.content, null);
		        }

		    } else {

		        if(m.img_64 != null){
		            msg = getMessage(m.name, m.photo, LEFT, m.content ,m.img_64);
		        }
		        else{
		            msg = getMessage(m.name, m.photo, LEFT, m.content, null);
		        }
		    }
		    break;
		case EVENT_OTHER:
		    if (m.name != PERSON_NAME) {
		        msg = getEventMessage(m.name + " " + m.content+" " )
		    } else {
		        msg = getEventMessage("You have" + " " + m.content+" " )
		    }
		    break;
	    }
	    insertMsg(msg, chatroom[0]);
	};

	ws.onclose = function (e) {
	    console.log(e)
	}
}
//send by click Send
send.onclick = function (e) {
    handleMessageEvent($("#msg").val())
}


//send by press Enter
text.onkeydown = function (e) {
    if (e.keyCode === 13) {
        handleMessageEvent($("#msg").val())
    }
};


//content to send
function handleMessageEvent(val) {
$("#msg").val("")
 if(val.trim() != ""){
    ws.send(JSON.stringify({
        "event": "message",
        "photo": PERSON_IMG,
        "name": PERSON_NAME,
        "content": val,
        "img_64" : res,
    }));
    var img = document.getElementById("image")
    img.remove()
    res ="";
    console.log(res)
    }
    else{
    alert("cannot send empty text!");
    }
}

function getEventMessage(msg) {
    var msg = `<div class="msg-left">${msg}</div>`
    return msg
}

function getMessage(name, img, side, text, img64) {
    const d = new Date()
    //   Simple solution for small apps
    var msg = `
    <div class="msg ${side}-msg">
    `
    if(img != null){
      msg +=`<div class="msg-img" style="background-image: url(${img})"></div>`
      }
	msg += `
      <div class="msg-bubble">
        <div class="msg-info">
          <div class="msg-info-name">${name}</div>
          <div class="msg-info-time">${d.getFullYear()}/${d.getMonth()}/${d.getDay()} ${d.getHours()}:${d.getMinutes()}</div>
        </div>

        <div class="msg-text">${text}`
        if(img64 != null){
        	msg += `<img src =${img64} id="return_image"/>`
        }
        msg +=`</div>
      </div>
    </div>
  `

    return msg;
}

function insertMsg(msg, domObj) {
    domObj.insertAdjacentHTML("beforeend", msg);
    domObj.scrollTop += 500;
}

function getRandomNum(min, max) {
    return Math.floor(Math.random() * (max - min + 1)) + min;
}
