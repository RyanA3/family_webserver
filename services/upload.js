require("dotenv").config();
const is_production = process.env.NODE_ENV === "production";
const files_dir = process.env.FILES_DIR;

const ImageFileMeta = require('../models/ImageFileMeta');
const mongoose = require('mongoose');

const multer = require('multer');

const storage = multer.diskStorage({
    destination: (req, file, cb) => {
        cb(null, files_dir);
    },
    filename: (req, file, cb) => {
        cb(null, Date.now() + '-' + file.originalname);
    }
});

const upload = multer({storage: storage});

module.exports = upload;