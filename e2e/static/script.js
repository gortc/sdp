
let configuration = {
    iceServers: [{
        urls: ["stun:stun.l.google.com:19302", "stun:stun1.l.google.com:19302"]
    }]
};

let pc = new RTCPeerConnection(configuration);
pc.createDataChannel('webrtchacks');

pc.createOffer(
    function (offer) {
        fetch("/sdp",
            {
                method: "POST",
                body: offer.sdp
            })
            .then(function (res) {
                return res.text();
            });
        pc.setLocalDescription(offer);
    },
    function (err) {
        console.error(err);
    }
);
