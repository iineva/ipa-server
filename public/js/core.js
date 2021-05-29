(function(exports) {

        dayjs.extend(window.dayjs_plugin_relativeTime)
        var lan = window.navigator.language
        if (lan.startsWith("zh")) {
            dayjs.locale("zh-cn")
        }

        // fetch with progress
        function fetch(url, opts = {}, onProgress) {
            return new Promise((res, rej) => {
                var xhr = new XMLHttpRequest()
                xhr.open(opts.method || 'get', url)
                for (var k in opts.headers || {})
                    xhr.setRequestHeader(k, opts.headers[k])
                xhr.onload = e => res(JSON.parse(e.target.responseText))
                xhr.onerror = rej
                if (xhr.upload && onProgress)
                    xhr.upload.onprogress = onProgress
                xhr.send(opts.body)
            });
        }

        function getApiUrl(path) {
            if (window.localStorage.getItem('ACCESS_KEY')) {
                return path + '?key=' + window.localStorage.getItem('ACCESS_KEY') + `&v=${parseInt(new Date().getTime() / 1000)}`
            }
            return path
        }

        // return true if is PC
        function isPC() {
            const Agents = ["Android", "iPhone", "SymbianOS", "Windows Phone", "iPad", "iPod"]
            for (let v = 0; v < Agents.length; v++) {
                if (window.navigator.userAgent.indexOf(Agents[v]) > 0) {
                    return false
                }
            }
            return true
        }

        function language() {
            return (navigator.language || navigator.browserLanguage)
        }

        // set locale for server
        document.cookie = `locale=${language()};`

        // localization string
        function langString(key) {
            const localStr = {
                'Download': {
                    'zh-cn': '下载'
                },
                'Upload Date: ': {
                    'zh-cn': '更新时间：'
                },
                'Add': {
                    'zh-cn': '添加'
                },
                'Upload Done!': {
                    'zh-cn': '上传成功！'
                },
                'Download and Install': {
                    'zh-cn': '下载安装'
                },
                'Beta': {
                    'zh-cn': '内测版'
                },
                'Current': {
                    'zh-cn': '当前'
                },
                'Channel': {
                    'zh-cn': '渠道'
                },
                'Delete': {
                    'zh-cn': '删除'
                },
                'Back to home?': {
                    'zh-cn': '是否返回首页？'
                },
                'Confirm to Delete?': {
                    'zh-cn': '确认删除？'
                },
                'Delete Success!': {
                    'zh-cn': '删除成功！'
                },
            }
            const lang = (localStr[key] || key)[language().toLowerCase()]
            return lang ? lang : key
        }

        // bytes to Human-readable string
        function sizeStr(size) {
            const K = 1024,
                M = 1024 * K,
                G = 1024 * M
            if (size > G) {
                return `${(size/G).toFixed(2)} GB`
            } else if (size > M) {
                return `${(size / M).toFixed(2)} MB`
            } else {
                return `${(size / K).toFixed(2)} KB`
            }
        }

        window.ipaInstall = function(event, plist) {
            event && event.stopPropagation()
            window.location.href = 'itms-services://?action=download-manifest&url=' + plist
        }

        window.goToLink = function(event, link) {
            event && event.stopPropagation()
            window.location.href = link
        }

        onInstallClick = function(row) {
            var needGoAppPage = !!(
                row.type === 0 ?
                (row.history || []).find(r => r.type === 1) :
                (row.history || []).find(r => r.type === 0)
            )
            if (needGoAppPage) {
                return `goToLink(null, '/app?id=${row.id}')`
            }

            if (row.type == 0) {
                return `ipaInstall(event, '${row.plist}')`
            }
            return `goToLink(event, '${row.pkg}')`
        }

        function createItem(row) {
            var icons = [row.type === 0 ? 'ios' : 'android'];
            (row.history || []).forEach(r => {
                if (r.type === 0 && icons.indexOf('ios') === -1) {
                    icons.push('ios')
                }
                if (r.type === 1 && icons.indexOf('android') === -1) {
                    icons.push('android')
                }
            });
            icons.sort().reverse()
            return `
      <a class='row' onclick="${`goToLink(event, '/app?id=${row.id}')`}">
        <img data-normal="${row.webIcon}" alt="">
        <div class="center">
          <div class="name">
            ${row.name}
            ${icons.map(t => `<img class="icon-tag ${t}" src="/img/${t}.svg">`).join('')}
            ${row.current ? `<span class="tag">${langString('Current')}</span>` : ''}
          </div>
          <div class="version">
            <span>${row.version}(Build ${row.build})</span>
            <span>${row.channel && IPA.langString('Channel') + ': '+row.channel || ''}</span>
          </div>
          <div class="date">${IPA.langString('Upload Date: ')}${dayjs(row.date).fromNow()}</div>
        </div>
        <div onclick="${onInstallClick(row)}" style="pointer-events:auto;" class="right">${IPA.langString('Download')}</div>
      </a>
    `
  }

  exports.IPA = {
    fetch: fetch,
    isPC: isPC(),
    langString: langString,
    sizeStr: sizeStr,
    createItem: createItem,
    getApiUrl: getApiUrl,
  }

})(window)