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

const datefns = require('date-fns');
const { format } = datefns;


function formatBytes(bytes) {

    if(bytes / 1000 < 1) return `${bytes} Bytes`;
    if(bytes / 1000000 < 1) return `${Math.round(bytes / 10) / 100} KB`;
    if(bytes / 1000000000 < 1) return `${Math.round(bytes / 10000) / 100} MB`;
    return `${Math.round(bytes / 10000000) / 100} GB`;

}


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

const page_size = 12;
var image_cards = pug.compileFile(`${root_directory}/views/components/image_cards.pug`)
router.get('/images/:page', async (req, res) => {

    var page = req.params.page;
    var sort = { created: 'desc' }
    var filter = {};
    if(req.query && req.query.sortBy) sort = { [req.query.sortBy]: req.query.sortDirection }

    if(req.query && req.query.showDuplicates && req.query.showDuplicates == "no" || (!req.query.showDuplicates)) {
        console.log("Not showing duplicates")
        filter = { 
            ...filter, 
            $or: [
                {duplicates: null},
                {duplicates: {$exists: false}},
                {duplicates: []}
            ]
        } 
        
    }   

    console.log(req.query);

    try {
        var metas = await ImageMeta.find(filter)
        .sort(sort)
        .limit(page_size)
        .skip(page * page_size)
        .select({ 
            _id: 1,
            file_size: 1,
            original_name: 1,
            created: 1,
            uploaded: 1,
            extension: 1
        })

        var displayInfo = [metas.length];

        //Format dates for display
        for(var i = 0; i < metas.length; i++) {
            displayInfo[i] = {
                created: format(new Date(metas[i].created), "MMM do y"),
                uploaded: format(new Date(metas[i].uploaded), "MMM do y"),
                file_size: formatBytes(Number(metas[i].file_size)),
                original_name: metas[i].original_name,
            }
        }


        console.log("Got images for client:\n " + metas);

        //imageNames = metas[0]._id.toHexString() + "." + metas[0].extension
        var imageNames = [metas.length]
        
        for(var i = 0; i < metas.length; i++) {
            imageNames[i] = metas[i]._id.toHexString();
            if(metas[i].extension) imageNames[i] += "." + metas[i].extension
        }
        console.log("\n\nGenerated names: " + imageNames + "\n\n")

        if(!is_production) image_cards = pug.compileFile(`${root_directory}/views/components/image_cards.pug`)
        res.status(200).send(image_cards({imageNames: imageNames, imageMetas: displayInfo}))


    } catch (e) {
        console.error(e)
    }

})

var pageChanger = pug.compileFile(`${root_directory}/views/components/page_changer.pug`)
router.get('/pageChanger/:page', (req, res) => {
    if(!is_production) pageChanger = pug.compileFile(`${root_directory}/views/components/page_changer.pug`)
    page = req.params.page;
    if(page < 0) page = 0
    res.status(200).send(pageChanger({page: page}))
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