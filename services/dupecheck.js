require("dotenv").config();
var exec = require('child_process').exec;


function processImages() {

    //exec("./dupecheck/dupecheck")
    //var file = new File("./dupecheck/dupecheck");
    //file.exec;
    exec("dupecheck", (err, stdout, stderr) => {
        console.log(err, stdout, stderr);
    })

}

module.exports = processImages