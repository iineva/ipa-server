(function(exports) {

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

        function isInAppBrowser() {
          /**
           * MicroMessenger => WeChat
           * QQ/ => QQ
           * AliApp => Taobao / Alipay
           */
            const Agents = ["MicroMessenger", "QQ/", "AliApp"]
            for (let v = 0; v < Agents.length; v++) {
                if (window.navigator.userAgent.indexOf(Agents[v]) > 0) {
                    return true
                }
            }
            return false
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
                "Click the button in the upper right corner, and then in the pop-up menu, click 'Open in Safari' to install.": {
                    'zh-cn': '点击右上角按钮，然后在弹出的菜单中，点击"在Safari中打开"，即可安装。'
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

        function createItem(row) {
            return `
      <a class='row' href="/app?id=${row.id}">
        <img data-normal="${row.webIcon}" alt="">
        <div class="center">
          <div class="name">${row.name}${row.current?`<span class="tag">${langString('Current')}</span>`:''}</div>
          <div class="version">
            <span>${row.version}(Build ${row.build})</span>
            <span>${row.channel && IPA.langString('Channel') + ': '+row.channel || ''}</span>
          </div>
          <div class="date">${IPA.langString('Upload Date: ')}${row.date}</div>
        </div>
        <div onclick="onClickInstall('${row.plist}')" class="right">${IPA.langString('Download')}</div>
      </a>
    `
  }

  exports.IPA = {
    fetch: fetch,
    isPC: isPC(),
    isInAppBrowser: isInAppBrowser(),
    langString: langString,
    sizeStr: sizeStr,
    createItem: createItem,
    getApiUrl: getApiUrl,
  }

})(window)