require("dotenv").config();
const is_production = process.env.NODE_ENV === "production";
const processing_dir = process.env.FILES_DIR + 'processing/';
const success_dir = process.env.FILES_DIR + 'images/';




function talkToGo() {
    // var lib = ffi.Library("./dupecheck/dupecheck.so", {
    //     'ProcessUploadedImages': ['string', ['string', 'string']]
    // })

    // var output = lib.ProcessUploadedImages(processing_dir, success_dir)
    // console.log("Node got output from go: ", output)
}

module.exports = talkToGo