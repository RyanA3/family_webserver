require("dotenv").config();
const is_production = process.env.NODE_ENV === "production";
const root_directory = process.env.ROOT_DIR;

const pug = require('pug');
const errMsg = pug.compileFile(`${root_directory}/views/components/error_message.pug`);

const express = require('express');
const router = express.Router();

const multer = require('multer');
const multi_upload = require('../services/upload/upload.js');



const uploadPage = pug.compileFile(`${root_directory}/views/pages/upload.pug`);
router.post('/upload', (req,res) => {
    multi_upload(req, res, (err) => {
        if(err instanceof multer.MulterError) {
            res.send(errMsg({name: err.name, stack: err.stack})).status(500).end();
            return;
        } else if(err) {
            if(err.name == 'ExtensionError') res.send(errMsg({name: err.name})).status(413).end();
            else res.send(errMsg({name: err.name})).status(500).end();
            return;
        }

        res.send(uploadPage()).status(200);
    })
})

module.exports = router;