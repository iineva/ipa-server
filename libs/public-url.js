const url = require('url')

module.exports = ctx => {
  const h = ctx.req.headers
  if (h.referer) {
    const u = url.parse(h.referer)
    return `${u.protocol}//${u.host}`
  } else {
    return `${h['x-forwarded-proto'] || 'http'}://${h.host}`
  }
}