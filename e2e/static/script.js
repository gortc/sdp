
let configuration = {
    iceServers: [{
        urls: ["stun:stun.l.google.com:19302", "stun:stun1.l.google.com:19302"]
    }]
};

let pc = new RTCPeerConnection(configuration);
pc.createDataChannel('webrtchacks');

pc.onicecandidate = function (event) {
    console.log(event);
};

pc.createOffer(
    function (offer) {
        fetch("/",
            {
                method: "POST",
                body: offer.sdp
            })
            .then(function (res) {
                return res.text();
            });
        pc.setLocalDescription(offer);
        console.log(offer.sdp);
    },
    function (err) {
        console.error(err);
    }
);
