require("dotenv").config();
const is_production = process.env.NODE_ENV === "production";
const files_dir = process.env.FILES_DIR + 'processing/';

const ImageFileMeta = require('../models/ImageMeta');
const mongoose = require('mongoose');

const multer = require('multer');
const allowedFileTypes = ["image/png", "image/jpg", "image/jpeg"]
const maxNumFiles = 1000;
const maxFileSize = 1 * 1024 * 1024 * 10; //10 MB



const storage = multer.diskStorage({
    destination: (req, file, cb) => {
        cb(null, files_dir);
    },
    filename: (req, file, cb) => {
        cb(null, Date.now() + '-' + file.originalname);
    }
});

//const upload = multer({storage: storage});

const multi_upload = multer({
    storage,
    limits: { fileSize: maxFileSize },
    fileFilter: (req, file, cb) => {
        if(allowedFileTypes.includes(file.mimetype)) {
            return cb(null, true);
        }

        const err = new Error("Only files of type ", allowedFileTypes, " are acceptable!");
        err.name = 'InvalidFileType';
        return cb(err);
    }
}).array('files', maxNumFiles);

module.exports = multi_upload;