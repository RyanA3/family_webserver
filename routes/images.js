require("dotenv").config();
const is_production = process.env.NODE_ENV === "production";
const root_directory = process.env.ROOT_DIR;

const pug = require('pug');

const express = require('express');
const router = express.Router();

const upload = require('../services/upload.js');



router.post('/upload', upload.single('file'), (req,res) => {
    res.send('<div>Got upload</div>');
})

router.get('/', (req,res) => {

}) 

module.exports = router;