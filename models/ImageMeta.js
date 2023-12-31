const mongoose = require('mongoose');

const ImageMeta = new mongoose.Schema({
    uploaded: {
        type: Date,
    },
    created: {
        type: Date
    },
    camera_make: {
        type: String
    },
    camera_model: {
        type: String
    },
    file_size: {
        type: Number
    },
    original_name: {
        type: String
    },
    extension: {
        type: String,
    },
    duplicates: [String]
})

module.exports = mongoose.model(process.env.MONGO_IMAGE_META_COLLECTION_NAME, ImageMeta)