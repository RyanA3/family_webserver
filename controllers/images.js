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

const DEFAULT_PAGE_SIZE = 12;
const MIN_PAGE_SIZE = 1;
const MAX_PAGE_SIZE = 100;
var image_cards = pug.compileFile(`${root_directory}/views/components/image_cards.pug`)
router.get('/images/:page', async (req, res) => {

    //Parse query and options from request
    var page = req.params.page;
    var pageSize = DEFAULT_PAGE_SIZE;
    var sort = { created: 'desc' }
    var filter = {};

    if(req.query && req.query.pageSize) {
        if(req.query.pageSize < MIN_PAGE_SIZE) pageSize = MIN_PAGE_SIZE;
        else if(req.query.pageSize > MAX_PAGE_SIZE) pageSize = MAX_PAGE_SIZE;
        else pageSize = req.query.pageSize;
    }
    if(req.query && req.query.sortBy) {
        sort = { 
            [req.query.sortBy]: req.query.sortDirection 
        }
    }
    if(req.query && req.query.showDuplicates && req.query.showDuplicates == "no" || (!req.query.showDuplicates)) {
        filter = { 
            ...filter, 
            $or: [
                {duplicates: null},
                {duplicates: {$exists: false}},
                {duplicates: []}
            ]
        }
    }

    if(req.query && req.query.filterAfter && req.query.filterAfter !== "null") {
        filter = {
            ...filter,
            created: {
                $gte: req.query.filterAfter
            }
        }
    }

    if(req.query && req.query.filterBefore && req.query.filterBefore !== "null") {
        filter = {
            ...filter,
            created: {
                $lte: req.query.filterBefore
            }
        }
    }

    if(req.query && req.query.filterCamera && req.query.filterCamera !== "null") {
        filter = {
            ...filter,
            $or: [
                { 
                    camera_make: {
                        $regex: req.query.filterCamera,
                    } 
                },
                { 
                    camera_model: {
                        $regex: req.query.filterCamera,
                    } 
                }
            ]
        }
    }

    if(req.query && req.query.filterOriginalName && req.query.filterOriginalName !== "null") {
        filter = {
            ...filter,
            original_name: {
                $regex: req.query.filterOriginalName,
            }
        }
    }

    //Do the query
    try {
        var metas = await ImageMeta.find(filter)
        .sort(sort)
        .limit(pageSize)
        .skip(page * pageSize)
        .select({ 
            _id: 1,
            file_size: 1,
            original_name: 1,
            created: 1,
            uploaded: 1,
            extension: 1,
            camera_make: 1,
            camera_model: 1,
        })

        //Return empty response if nothing was found
        if(metas.length == 0) {
            res.status(200).send("<div id=\"imageCards\"><div>Nothing Found</div></div>");
            return;
        }

        var displayInfo = [metas.length];

        //Format fields for display
        for(var i = 0; i < metas.length; i++) {
            displayInfo[i] = {
                created: format(new Date(metas[i].created), "MMM do y"),
                uploaded: format(new Date(metas[i].uploaded), "MMM do y"),
                file_size: formatBytes(Number(metas[i].file_size)),
                original_name: metas[i].original_name,
                camera: metas[i].camera_make + " " + metas[i].camera_model,
            }
        }

        var imageNames = [metas.length]
        
        for(var i = 0; i < metas.length; i++) {
            imageNames[i] = metas[i]._id.toHexString();
            if(metas[i].extension) imageNames[i] += "." + metas[i].extension
        }

        if(!is_production) image_cards = pug.compileFile(`${root_directory}/views/components/image_cards.pug`)
        res.status(200).send(image_cards({imageNames: imageNames, imageMetas: displayInfo}))

    } catch (e) {
        console.error("ERROR OCCURRED WHILE QUERYING IMAGES", e)
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