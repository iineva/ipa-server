const plist = require('simple-plist')
const DecompressZip = require('decompress-zip')
const config = require('../config')
const fs = require('fs-extra')
const path = require('path')
const moment = require('moment')
const pngdefry = require('pngdefry')

const uploadDir = config.uploadDir

fs.ensureDir(uploadDir)

// store data
const appListFile = path.join(uploadDir, 'appList.json')
const appList = []

// init appList
if (fs.pathExistsSync(appListFile)) {
  const list = fs.readJsonSync(appListFile)
  list.map(row => appList.push(row))
}

const iconPath = (publicURL, row) => {
  if (row.noneIcon) {
    return 'img/default.png'
  } else {
    return `${row.identifier}/${row.id}/icon.png`
  }
}

const itemInfo = (row, publicURL) => Object.assign({}, row, {
  ipa: `${config.publicURL || publicURL}/${row.identifier}/${row.id}/ipa.ipa`,
  icon: `${config.publicURL || publicURL}/${iconPath(publicURL, row)}`,
  plist: `${config.publicURL || publicURL}/plist/${row.id}.plist`,
  webIcon: `/${iconPath(publicURL, row)}`, // to display on web
  date: moment(row.date).fromNow(),
})

const list = (publicURL) => {

  const backList = []

  appList.map(row => itemInfo(row, publicURL)).map(row => {
    let app = backList.find(r => r.identifier === row.identifier)
    if (!app) {
      app = row
      backList.push(app)
    }
    app.history = app.history || []
    app.history.push(Object.assign({}, row, {history: undefined, current: row.id===app.id}))
  })

  return backList
}

const decompress = (opt) => new Promise((resolve, reject) => {
  const unzipper = new DecompressZip(opt.file)
  unzipper.on('error', reject)
  unzipper.on('extract', resolve)
  unzipper.extract(opt)
})

const fixPNG = (input, output) => new Promise((resolve, reject) => {
  pngdefry(input, output, (err) => err ? reject(err) : resolve())
})

const add = async (file) => {

  const tmpDir = '/tmp/cn.ineva.upload/unzip-tmp' // temp dir
  let plistFile, iconFiles = []

  // unzip files
  const newIconRegular = /Payload\/.*\.app\/AppIcon-?(\d+(\.\d+)?)x(\d+(\.\d+)?)(@\dx)?.*\.png$/
  const oldIconRegular = /Payload\/.*\.app\/Icon-?(\d+(\.\d+)?)?.png$/
  await fs.remove(tmpDir)
  await decompress({
    file: file,
    path: tmpDir,
    filter: (file) => {
      if (file.path.endsWith('.app/Info.plist')) {
        plistFile = file
        return true
      } else if (
        file.path.match(newIconRegular) ||
        file.path.match(oldIconRegular)
      ) {
        iconFiles.push(file)
        return true
      } else {
        return false
      }
    }
  })

  // select max size icon
  let iconFile, maxSize = 0
  iconFiles.forEach(file => {
    let size = 0
    if (file.path.match(oldIconRegular)) {
      // parse old icons
      const arr = path.basename(file.path, '.png').split('-')
      if (arr.length === 2) {
        size = Number(arr[1])
      } else {
        size = 160
      }
    } else {
      // parse new icons
      size = Number(path.basename(file.path, '.png').split('@')[0].split('x')[1].split('~')[0])
      if (file.path.indexOf('@2x') !== -1) {
        size *= 2
      } else if (file.path.indexOf('@3x') !== -1) {
        size *= 3
      }
    }
    if (size > maxSize) {
      maxSize = size
      iconFile = file
    }
  })

  // parse plist
  const info = plist.readFileSync(path.join(tmpDir, plistFile.path))
  const app = {
    id: path.basename(file, '.ipa'),
    name: info['CFBundleDisplayName'] || info['CFBundleName'] || info['CFBundleExecutable'],
    version: info['CFBundleShortVersionString'],
    identifier: info['CFBundleIdentifier'],
    build: info['CFBundleVersion'],
    date: new Date(),
    size: (await fs.lstat(file)).size,
    noneIcon: !iconFile,
    channel: info['channel'],
  }
  appList.unshift(app)
  await fs.writeJson(appListFile, appList)

  // save files to target dir
  // TODO: upload dir configable
  const targetDir = path.join(uploadDir, app.identifier, app.id)
  await fs.move(file, path.join(targetDir, 'ipa.ipa'))
  if (iconFile) {
    try {
      await fixPNG(path.join(tmpDir, iconFile.path), path.join(targetDir, 'icon.png'))
    } catch (err) {
      await fs.move(path.join(tmpDir, iconFile.path), path.join(targetDir, 'icon.png'))
    }
  }

  // delete temp files
  await fs.remove(tmpDir)
}

const find = (id, publicURL) => {
  const row = itemInfo(appList.find(row => row.id === id), publicURL)
  if (!row) { return {} }

  row.history = appList.filter(r => r.identifier === row.identifier).map(r => Object.assign({}, itemInfo(r, publicURL), {
    current: r.id === row.id,
  }))

  return row
}

module.exports = {
  list,
  find,
  add,
}
