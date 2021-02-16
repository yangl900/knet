var express = require('express');
var router = express.Router();

var clients = new Map()
var ipports = new Map()

router.get('/', function(req, res, next) {
    ip = req.headers['x-forwarded-for'] || req.socket.remoteAddress
    console.log("IP " + ip)
    ip = ip.replace("::ffff:", "")
    addr = ip.split(':')[0]
    port = ip.split(':')[1] || req.socket.remotePort
    fullip = addr + ":" + port

    console.log(fullip)

    if (!ipports.has(fullip)) {
        ipports.set(fullip, 1)

        if (clients.has(addr)) {
            curr = clients.get(addr) || 0
            clients.set(addr, curr + 1)
            
            console.log("append connection for addrss " + addr)
        } else {
            console.log("add new addrss " + addr)
            clients.set(addr, 1)
        }

        setTimeout(function() {
            res.send("Request from new connection on " + fullip);
            curr = clients.get(addr) || 0
            clients.set(addr, curr - 1)
            ipports.delete(fullip)

            if (clients.get(addr) <= 0) {
                clients.delete(addr)
            }
        },
        60 * 1000);
    } else {
        res.send("Requst from established connection " + fullip);
    }
  });

module.exports = {
    register: function() { return router },
    clients: function() { return Object.fromEntries(clients) }
}
