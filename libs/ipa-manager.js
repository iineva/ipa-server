const plist = require('simple-plist')
const DecompressZip = require('decompress-zip')
const config = require('../config')
const fs = require('fs-extra')
const path = require('path')
const moment = require('moment')
const pngdefry = require('pngdefry')

// 数据保存
const appListFile = path.resolve(__dirname, '../upload/appList.json')
const appList = []

// 初始化appList
if (fs.pathExistsSync(appListFile)) {
  const list = fs.readJsonSync(appListFile)
  list.map(row => appList.push(row))
}

const list = () => appList.map(row => Object.assign({}, row, {
  ipa: `${config.publicURL}/${row.identifier}/${row.id}/ipa.ipa`,
  icon: `${config.publicURL}/${row.identifier}/${row.id}/icon.png`,
  plist: `${config.publicURL}/plist/${row.id}.plist`,
  webIcon: `/${row.identifier}/${row.id}/icon.png`, // 用于web页面显示
  date: moment(row.date).fromNow(),
}))

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

  const tmpDir = '/tmp/cn.ineva.upload/unzip-tmp' // 缓存目录
  let plistFile, iconFiles = []

  // unzip files
  const newIconRegular = /Payload\/\w*\.app\/AppIcon(\d+(\.\d+)?)x(\d+(\.\d+)?)(@\dx)?.*\.png$/
  const oldIconRegular = /Payload\/\w*\.app\/Icon(-\d+(\.\d+)?)?.png$/
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
  console.log(iconFiles.map(row => row.path))
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

  // 解析plist
  const info = plist.readFileSync(path.join(tmpDir, plistFile.path))
  const app = {
    id: path.basename(file, '.ipa'),
    name: info['CFBundleDisplayName'] || info['CFBundleName'] || info['CFBundleExecutable'],
    version: info['CFBundleShortVersionString'],
    identifier: info['CFBundleIdentifier'],
    build: info['CFBundleVersion'],
    date: new Date(),
    size: (await fs.lstat(file)).size,
  }
  appList.unshift(app)
  await fs.writeJson(appListFile, appList)

  // 保存相关文件到指定目录
  // TODO: 设置upload目录
  const targetDir = path.resolve(__dirname, '../upload', app.identifier, app.id)
  await fs.move(file, path.join(targetDir, 'ipa.ipa'))
  try {
    await fixPNG(path.join(tmpDir, iconFile.path), path.join(targetDir, 'icon.png'))
  } catch (err) {
    await fs.move(path.join(tmpDir, iconFile.path), path.join(targetDir, 'icon.png'))
  }

  // 删除无用文件
  await fs.remove(tmpDir)
}

const find = id => list().find(row => row.id === id)

module.exports = {
  list,
  find,
  add,
}
