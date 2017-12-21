module.exports = {
  debug: process.env.NODE_ENV !== 'production',
  host: process.env.HOST || '0.0.0.0',
  port: process.env.PORT || 8080,
  publicURL: process.env.PUBLIC_URL || 'https://ipa.ineva.cn',
}
