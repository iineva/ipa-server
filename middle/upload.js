const fs = require('fs-extra')
const path = require('path')
const Busboy = require('busboy')
const randomstring = require('randomstring')
const mime = require('mime-types')
const url = require('url')

const receiveFiles = async (req, onFile)=>{
  return new Promise((resolve, reject)=>{
    if (req.method === 'POST') {
      const busboy = new Busboy({ headers: req.headers });
      busboy.on('file', function(fieldname, file, filename, encoding, mimetype) {
        onFile({ fieldname, file, filename, encoding, mimetype })
      })
      busboy.on('finish', resolve)
      busboy.on('error', reject)
      req.pipe(busboy)
    } else {
      resolve()
    }
  })
}

const saveStream = (stream, pathName)=>{
  return new Promise((resolve, reject)=> {
    const fileWriteStream = fs.createWriteStream(pathName)
    stream.pipe(fileWriteStream)
  })
}

module.exports = (opt = {}, next) => {
  opt.tempDir = opt.tempDir || '/tmp/cn.ineva.upload' // default temp dir
  opt.defExt = opt.defExt || 'jpg' // default ext
  return async (ctx) => {
    const files = []
    await receiveFiles(ctx.req, async (row)=>{
      const ext = opt.defExt || mime.extension(row.mimetype)
      const filename = randomstring.generate({length: 16, charset: 'hex'}) + (ext?`.${ext}`:'')
      await fs.ensureDir(path.join(opt.tempDir))
      saveStream(row.file, path.join(opt.tempDir, filename))
      files.push(path.join(opt.tempDir, filename))
    })
    await next(ctx, files)
  }
}
