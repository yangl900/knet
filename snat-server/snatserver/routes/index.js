var express = require('express');
var hook = require('./hook');
var router = express.Router();

/* GET home page. */
router.get('/', function(req, res, next) {
  res.render('index', { title: 'Http Connections Test Site', clients: hook.clients() });
});

module.exports = router;
