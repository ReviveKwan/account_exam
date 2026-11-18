#!/usr/bin/env node
/**
 * 小程序 CI 预览脚本：生成预览二维码（无需打开微信开发者工具）
 *
 * 用法：
 *   export WX_PRIVATE_KEY_PATH=/path/to/private.key
 *   node scripts/preview.js
 *   生成的二维码保存在 scripts/output/preview-qrcode.png，用微信扫码即可真机预览
 */
const path = require('path')
const fs = require('fs')
const ci = require('miniprogram-ci')

const APPID = 'wxe161a5d9e2c0f148'
const PROJECT_ROOT = path.resolve(__dirname, '..', 'miniprogram')
const OUT_DIR = path.resolve(__dirname, 'output')
const KEY_PATH = process.env.WX_PRIVATE_KEY_PATH

function fail(msg) {
  console.error('[preview] 失败:', msg)
  process.exit(1)
}

if (!KEY_PATH || !fs.existsSync(KEY_PATH || '')) {
  fail('请先设置 WX_PRIVATE_KEY_PATH 指向有效的代码上传私钥文件')
}

async function main() {
  fs.mkdirSync(OUT_DIR, { recursive: true })
  const privateKey = fs.readFileSync(KEY_PATH, 'utf8')
  const project = new ci.Project({
    appid: APPID,
    type: 'miniProgram',
    projectPath: PROJECT_ROOT,
    privateKey,
    ignores: ['node_modules/**/*', '.git/**/*'],
  })

  try {
    const qrcodePath = path.join(OUT_DIR, 'preview-qrcode.png')
    const result = await ci.preview({
      project,
      desc: 'CI 预览',
      setting: { es6: true, minify: true },
      qrcodeFormat: 'image',
      qrcodeOutputDest: qrcodePath,
      onProgressUpdate: () => {},
    })
    console.log('[preview] 预览二维码已生成:', qrcodePath)
    console.log('[preview] 用微信扫码即可真机预览（有效期约 25 分钟）')
    if (result) console.log('[preview] 结果:', JSON.stringify(result, null, 2))
  } catch (err) {
    fail(err && err.message ? err.message : String(err))
  }
}

main()
