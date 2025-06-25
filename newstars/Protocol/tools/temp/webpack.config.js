var path = require('path');
module.exports = {
    entry:'./main.js',
    output:{
        filename:'../js/ProtoLib.js',
        path:path.resolve(__dirname, './'),
        library:'ProtoLib'
    }
}
