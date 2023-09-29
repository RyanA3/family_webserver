require("dotenv").config();
const is_production = process.env.NODE_ENV === "production";
const root_directory = process.env.ROOT_DIR;

const pug = require('pug');
const errMsg = pug.compileFile(`${root_directory}/views/components/error_message.pug`);

const express = require('express');
const router = express.Router();

const multer = require('multer');
const multi_upload = require('../services/upload.js');
const processImages = require("../services/dupecheck.js");

const ImageMeta = require('../models/ImageMeta.js')



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

        processImages();
        res.send(uploadPage()).status(200);
    })

})

const page_size = 3;
var image_cards = pug.compileFile(`${root_directory}/views/components/image_cards.pug`)
router.get('/images/:page', async (req, res) => {

    var page = req.params.page;

    var sort = { created: 'asc' }
    if(req.body && req.body.sort) sort = req.body.sort;

    try {
        var metas = await ImageMeta.find({})
        .sort(sort)
        .limit(page_size)
        .skip(page * page_size)
        .select({ 
            _id: 1, 
            extension: 1
        })

        console.log("Got images for client:\n " + metas);

        //imageNames = metas[0]._id.toHexString() + "." + metas[0].extension
        var imageNames = [metas.length]
        
        for(var i = 0; i < metas.length; i++) {
            imageNames[i] = metas[i]._id.toHexString();
            if(metas[i].extension) imageNames[i] += "." + metas[i].extension
        }
        console.log("\n\nGenerated names: " + imageNames + "\n\n")

        if(!is_production) image_cards = pug.compileFile(`${root_directory}/views/components/image_cards.pug`)
        res.status(200).send(image_cards({imageNames: imageNames}))


    } catch (e) {
        console.error(e)
    }

})

var pageChanger = pug.compileFile(`${root_directory}/views/components/page_changer.pug`)
router.get('/pageChanger/:page', (req, res) => {
    if(!is_production) pageChanger = pug.compileFile(`${root_directory}/views/components/page_changer.pug`)
    console.log('Sent page changed:\n', pageChanger({page: req.params.page}))
    res.status(200).send(pageChanger({page: req.params.page}))
})

var fullImageCard = pug.compileFile(`${root_directory}/views/components/full_image_card.pug`)

router.get('/full/nofull', (req, res) => {
    res.status(200).send('<div id="fullImageModal"></div>')
})

router.get('/full/:imageName', (req, res) => {

    var imgName = req.params.imageName;

    if(!is_production) fullImageCard = pug.compileFile(`${root_directory}/views/components/full_image_card.pug`)
    res.status(200).send(fullImageCard({imgName: imgName}))

})



// router.get('/image/:id', (req, res) => {

// })

// router.get('/preview/:id', (req, res) => {

// })

module.exports = router;