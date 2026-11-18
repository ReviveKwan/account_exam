#!/usr/bin/env node
/**
 * 小程序 CI 上传脚本（使用 miniprogram-ci，无需打开微信开发者工具）
 *
 * 前置条件：
 *   1. 微信公众平台后台（mp.weixin.qq.com）→ 开发管理 → 开发设置
 *      → 小程序代码上传 → 生成/下载「代码上传私钥」(private.key)
 *   2. 同一页面把本机公网 IP 加入「IP 白名单」
 *      （当前出口 IP 见 README 或运行 `curl https://ifconfig.me/ip`）
 *
 * 用法：
 *   export WX_PRIVATE_KEY_PATH=/path/to/private.key
 *   node scripts/upload.js [版本备注]
 */
const path = require('path')
const fs = require('fs')
const ci = require('miniprogram-ci')

const APPID = 'wxe161a5d9e2c0f148'
const PROJECT_ROOT = path.resolve(__dirname, '..', 'miniprogram')
const KEY_PATH = process.env.WX_PRIVATE_KEY_PATH

function fail(msg) {
  console.error('[upload] 失败:', msg)
  process.exit(1)
}

if (!KEY_PATH) {
  fail('请先设置环境变量 WX_PRIVATE_KEY_PATH 指向代码上传私钥文件')
}
if (!fs.existsSync(KEY_PATH)) {
  fail(`私钥文件不存在: ${KEY_PATH}`)
}
if (!fs.existsSync(path.join(PROJECT_ROOT, 'project.config.json'))) {
  fail(`找不到项目配置: ${PROJECT_ROOT}/project.config.json`)
}

const desc = process.argv[2] || `CI上传 ${new Date().toISOString().slice(0, 16).replace('T', ' ')}`

async function main() {
  console.log('[upload] AppID:', APPID)
  console.log('[upload] 项目目录:', PROJECT_ROOT)
  console.log('[upload] 版本备注:', desc)

  const privateKey = fs.readFileSync(KEY_PATH, 'utf8')
  const project = new ci.Project({
    appid: APPID,
    type: 'miniProgram',
    projectPath: PROJECT_ROOT,
    privateKey,
    ignores: ['node_modules/**/*', '.git/**/*'],
  })

  try {
    const result = await ci.upload({
      project,
      version: '1.0.0',
      desc,
      setting: {
        es6: true,
        minify: true,
        autoPrefixWXSS: true,
      },
      onProgressUpdate: (task) => {
        if (task && task._id === 'MINIPROGRAM_UPLOAD') {
          console.log(`[upload] ${task.status || task.msg || ''}`)
        }
      },
    })
    console.log('[upload] 上传成功 ✅')
    console.log('[upload] 结果:', JSON.stringify(result, null, 2))
    console.log('[upload] 下一步：到微信公众平台 → 版本管理 → 选择该版本 → 提交审核')
  } catch (err) {
    fail(err && err.message ? err.message : String(err))
  }
}

main()
