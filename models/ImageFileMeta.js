const mongoose = require('mongoose');

const ImageFileMeta = new mongoose.Schema({
    uploadDate: {
        type: Date,
        required: true,
        default: new Date()
    },
    creationDate: {
        type: Date
    },
    format: {
        type: String,
        required: true,
    }
})

module.exports = mongoose.model("ImageFileMeta", ImageFileMeta)